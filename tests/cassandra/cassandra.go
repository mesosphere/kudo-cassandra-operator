package cassandra

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	log "github.com/sirupsen/logrus"
)

const (
	operatorYamlFilePath = "../../../operator/operator.yaml"
	// https://regex101.com/r/Ly7O1x/3/
	semVerRegexp = `(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
)

// OverrideOperatorVersion rewrites `operatorVersion` in operator.yaml. This is
// currently necessary in integration tests that run upgrades to make it
// possible to upgrade to a version based on the local filesystem even when the
// `operatorVersion` is the same as the one being upgraded from.
//
// Returns the operator version parsed before overriding and the operator
// version written.
func OverrideOperatorVersion(
	desiredOperatorVersion string,
) (string, string, error) {
	operatorYamlBytes, err := ioutil.ReadFile(operatorYamlFilePath)
	if err != nil {
		log.Errorf("Failed to read '%s': %v", operatorYamlFilePath, err)
		return "", "", err
	}

	scanner := bufio.NewScanner(bytes.NewReader(operatorYamlBytes))

	operatorVersionLineRegexp := fmt.Sprintf(
		`^operatorVersion:\s*"(%s)"$`, semVerRegexp,
	)
	operatorVersionLineLinePattern := regexp.MustCompile(operatorVersionLineRegexp)

	var operatorVersion string
	for scanner.Scan() {
		text := scanner.Text()
		match := operatorVersionLineLinePattern.FindStringSubmatch(text)
		if len(match) > 0 {
			operatorVersion = match[1]
		}
	}

	if operatorVersion == "" {
		errorMessage := fmt.Sprintf(
			"Failed to parse operatorVersion from '%s'", operatorYamlFilePath,
		)
		log.Error(errorMessage)
		return "", "", errors.New(errorMessage)
	}

	updatedOperatorYamlBytes := bytes.Replace(
		operatorYamlBytes,
		[]byte(operatorVersion),
		[]byte(desiredOperatorVersion),
		-1,
	)

	err = ioutil.WriteFile(operatorYamlFilePath, updatedOperatorYamlBytes, 0644)

	if err != nil {
		log.Errorf("Failed to write updated '%s': %v", operatorYamlFilePath, err)
		return "", "", err
	}

	log.Infof(
		"Overrode operatorVersion from '%s' to '%s'",
		operatorVersion,
		desiredOperatorVersion,
	)

	return operatorVersion, desiredOperatorVersion, nil
}

func firstPodName(instance kudo.Instance) (string, error) {
	if instance.Spec.Parameters["NODE_TOPOLOGY"] != "" {
		topology, err := TopologyFromYaml(instance.Spec.Parameters["NODE_TOPOLOGY"])
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal topology: %v", err)
		}

		return fmt.Sprintf("%s-%s-%s-%d", instance.Name, topology[0].Datacenter, "node", 0), nil
	} else {
		return fmt.Sprintf("%s-%s-%d", instance.Name, "node", 0), nil
	}
}

func Nodes(client client.Client, instance kudo.Instance) ([]map[string]string, error) {
	podName, err := firstPodName(instance)
	if err != nil {
		return nil, err
	}

	log.Infof("Get Node Status from %s", podName)

	var stdout strings.Builder

	cmd := cmd.New("nodetool").
		WithArguments("status").
		WithStdout(&stdout)

	pod, err := kubernetes.GetPod(client, podName, instance.Namespace)
	if err != nil {
		return nil, err
	}

	err = pod.ContainerExec("cassandra", cmd)
	if err != nil {
		return nil, err
	}
	// Example output:
	//	cassandra@cassandra-instance-us-west-2a-rac1-node-0:/$ nodetool status
	//	Datacenter: us-west-2a
	//	======================
	//	Status=Up/Down
	//	|/ State=Normal/Leaving/Joining/Moving
	//	--  Address          Load       Tokens       Owns (effective)  Host ID                               Rack
	//	UN  192.168.197.82   118.01 KiB  256          54.6%             0d0afae0-77be-44eb-b756-7ed47521592a  rac1
	//	UN  192.168.217.87   103.7 KiB  256          48.4%             1a546b75-c83b-4482-97e1-3e9621664a23  rac1
	//	Datacenter: us-west-2b
	//	======================
	//	Status=Up/Down
	//	|/ State=Normal/Leaving/Joining/Moving
	//	--  Address          Load       Tokens       Owns (effective)  Host ID                               Rack
	//	UN  192.168.8.141    117.92 KiB  256          48.0%             cf0c4d4f-e41e-43cb-a489-55095b62ca98  rac1
	//	UN  192.168.211.212  150.87 KiB  256          49.0%             7309951f-4eaa-4f4a-b526-632f975bc8ea  rac1

	// Datacenter: dc1
	dcRegexp := `^Datacenter:\s+(.*)$`
	dcLinePattern := regexp.MustCompile(dcRegexp)

	// --  Address         Load        Tokens  Owns   Host ID                               Rack
	// UN  192.168.196.13  105.29 KiB  256     68.8%  440b2d75-059c-444a-ab01-9cea29b387d8  rack1
	nodeRegexp := `^(\w{2})\s+([\w\.]+)\s+([\w\.]+\s\w+)\s+(\d+)\s+([\d\.]+%)\s+([\w-]+)\s+(.*)$`
	nodeLinePattern := regexp.MustCompile(nodeRegexp)

	var nodes []map[string]string

	scanner := bufio.NewScanner(strings.NewReader(stdout.String()))

	datacenter := ""

	for scanner.Scan() {
		dcMatch := dcLinePattern.FindStringSubmatch(scanner.Text())
		if dcMatch != nil {
			datacenter = dcMatch[1]
		}

		nodeMatch := nodeLinePattern.FindStringSubmatch(scanner.Text())
		if nodeMatch != nil {
			nodes = append(nodes, map[string]string{
				"status":     nodeMatch[1],
				"address":    nodeMatch[2],
				"load":       nodeMatch[3],
				"tokens":     nodeMatch[4],
				"owns":       nodeMatch[5],
				"host_id":    nodeMatch[6],
				"rack":       nodeMatch[7],
				"datacenter": datacenter,
			})
		}
	}

	return nodes, nil
}

// Cqlsh Wrapper to run cql commands in the cqlsh cli of cassandra 0th node
func Cqlsh(client client.Client, instance kudo.Instance, cql string) (string, error) {
	podName, err := firstPodName(instance)
	if err != nil {
		return "", err
	}

	log.Infof("Run CqlSh on %s", podName)

	var stdout strings.Builder
	var stderr strings.Builder

	cmd := cmd.New("cqlsh").
		WithArguments(fmt.Sprintf("--execute=%s", cql)).
		WithStdout(&stdout).
		WithStderr(&stderr)

	pod, err := kubernetes.GetPod(client, podName, instance.Namespace)
	if err != nil {
		return "", err
	}

	err = pod.ContainerExec("cassandra", cmd)
	if err != nil {
		log.Errorf("StdErr of failed CqlSh: %v", stderr.String())
		return "", err
	}

	return stdout.String(), nil
}

func Uninstall(client client.Client, operator kudo.Operator) error {
	// This wait is necessary to avoid tickling an issue in stateful set controller,
	// which gets stuck with a pod but no PVC when KUDO is quick to process the instance delete
	// (and create by subsequent test).
	if err := operator.UninstallWaitForDeletion(5 * time.Minute); err != nil {
		return err
	}

	pvcs, err := kubernetes.ListPersistentVolumeClaims(client, operator.Instance.Namespace)
	if err != nil {
		return err
	}

	for _, pvc := range pvcs {
		err := pvc.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func ClusterConfiguration(client client.Client, instance kudo.Instance) (map[string]string, error) {
	return configurationFromNodeLogs(
		client,
		instance,
		"org.apache.cassandra.config.Config - Node configuration:\\[(.+)\\]",
		";")
}

func NodeJVMOptions(client client.Client, instance kudo.Instance) (map[string]string, error) {
	return configurationFromNodeLogs(
		client,
		instance,
		"o.a.c.service.CassandraDaemon - JVM Arguments: \\[(.+)\\]",
		",")
}

func configurationFromNodeLogs(
	client client.Client,
	instance kudo.Instance,
	regex string,
	separator string) (map[string]string, error) {
	podName := fmt.Sprintf("%s-%s-%d", instance.Name, "node", 0)

	pod, err := kubernetes.GetPod(client, podName, instance.Namespace)
	if err != nil {
		return nil, err
	}

	logs, err := pod.ContainerLogs("cassandra")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(logs))
	configurationLinePattern := regexp.MustCompile(regex)

	var configurationLine string
	for scanner.Scan() {
		match := configurationLinePattern.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			configurationLine = match[1]
		}
	}

	configuration := make(map[string]string)

	for _, kv := range strings.Split(configurationLine, separator) {
		parts := strings.Split(strings.TrimSpace(kv), "=")
		if len(parts) == 2 {
			configuration[parts[0]] = parts[1]
		}
	}

	return configuration, nil
}
