package cassandra

import (
	"bufio"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	kudo "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

func ClusterConfiguration(
	namespaceName string, instanceName string,
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
	configurationLinePattern := regexp.MustCompile(
		"org.apache.cassandra.config.Config - Node configuration:\\[(.+)\\]",
	)
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

	for _, kv := range strings.Split(configurationLine, ";") {
		parts := strings.Split(strings.TrimSpace(kv), "=")
		if len(parts) == 2 {
			configuration[parts[0]] = parts[1]
		}
	}

	return configuration, nil
}