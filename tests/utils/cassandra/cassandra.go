package cassandra

import (
	"bufio"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

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
