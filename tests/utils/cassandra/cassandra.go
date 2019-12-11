package cassandra

import (
	"bufio"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

// ClusterConfiguration TODO function comment.
func ClusterConfiguration(
	namespaceName string, instanceName string,
) (map[string]string, error) {
	return getConfigurationFromNodeLogs(
		namespaceName,
		instanceName,
		"org.apache.cassandra.config.Config - Node configuration:\\[(.+)\\]",
		";",
	)
}

// NodeJvmOptions TODO function comment.
func NodeJvmOptions(
	namespaceName string, instanceName string,
) (map[string]string, error) {
	return getConfigurationFromNodeLogs(
		namespaceName,
		instanceName,
		"o.a.c.service.CassandraDaemon - JVM Arguments: \\[(.+)\\]",
		",",
	)
}

// Nodes TODO function comment.
func Nodes(
	namespaceName string, instanceName string,
) ([]map[string]string, error) {
	stdout, _ := kudo.ExecInPodContainer(
		namespaceName,
		instanceName,
		"node",
		0,
		"cassandra",
		[]string{"nodetool", "status"},
	)

	scanner := bufio.NewScanner(stdout)

	// --  Address         Load        Tokens  Owns   Host ID                               Rack
	// UN  192.168.196.13  105.29 KiB  256     68.8%  440b2d75-059c-444a-ab01-9cea29b387d8  rack1
	nodeRegexp := "^(\\w{2})\\s+([\\w\\.]+)\\s+([\\w\\.]+\\s\\w+)\\s+(\\d+)\\s+([\\d\\.]+%)\\s+([\\w-]+)\\s+(\\w+)$"
	nodeLinePattern := regexp.MustCompile(nodeRegexp)

	var nodes []map[string]string
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

func getConfigurationFromNodeLogs(
	namespaceName string,
	instanceName string,
	regexpr string,
	separator string,
) (map[string]string, error) {
	configuration := make(map[string]string)

	logs, err := kudo.GetPodContainerLogs(
		namespaceName, instanceName, "node", 0, "cassandra",
	)
	if err != nil {
		log.Errorf(
			"Error getting Cassandra node logs (instance='%s', namespace='%s'): %s",
			instanceName, namespaceName, err,
		)
		return configuration, err
	}

	scanner := bufio.NewScanner(logs)
	configurationLinePattern := regexp.MustCompile(regexpr)
	var configurationLine string
	for scanner.Scan() {
		match := configurationLinePattern.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			configurationLine = match[1]
		}
	}

	if configurationLine == "" {
		log.Warnf(
			"Couldn't find configuration line in Cassandra node logs "+
				"(instance='%s', namespace='%s'): %s",
			instanceName, namespaceName, err,
		)
		return configuration, err
	}

	for _, kv := range strings.Split(configurationLine, separator) {
		parts := strings.Split(strings.TrimSpace(kv), "=")
		if len(parts) == 2 {
			configuration[parts[0]] = parts[1]
		}
	}

	return configuration, nil
}
