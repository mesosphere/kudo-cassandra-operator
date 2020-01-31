package cassandra

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	log "github.com/sirupsen/logrus"
)

const (
	operatorYamlFilePath = "../../operator/operator.yaml"
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

func Nodes(client client.Client, instance kudo.Instance) ([]map[string]string, error) {
	podName := fmt.Sprintf("%s-%s-%d", instance.Name, "node", 0)

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

	// --  Address         Load        Tokens  Owns   Host ID                               Rack
	// UN  192.168.196.13  105.29 KiB  256     68.8%  440b2d75-059c-444a-ab01-9cea29b387d8  rack1
	nodeRegexp := "^(\\w{2})\\s+([\\w\\.]+)\\s+([\\w\\.]+\\s\\w+)\\s+(\\d+)\\s+([\\d\\.]+%)\\s+([\\w-]+)\\s+(\\w+)$"
	nodeLinePattern := regexp.MustCompile(nodeRegexp)

	var nodes []map[string]string

	scanner := bufio.NewScanner(strings.NewReader(stdout.String()))

	for scanner.Scan() {
		match := nodeLinePattern.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			nodes = append(nodes, map[string]string{
				"status":  match[1],
				"address": match[2],
				"load":    match[3],
				"tokens":  match[4],
				"owns":    match[5],
				"host_id": match[6],
				"rack":    match[7],
			})
		}
	}

	return nodes, nil
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
