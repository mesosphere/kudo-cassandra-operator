package suites

import (
	"fmt"
	"os"
	"time"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

func AssertNumberOfCassandraNodes(client client.Client, op kudo.Operator, nodeCount int) {
	Eventually(func() int {
		nodes, err := cassandra.Nodes(client, op.Instance)
		Expect(err).To(BeNil())

		upNormalNodes := 0
		for _, n := range nodes {
			if n["status"] == "UN" {
				upNormalNodes++
			} else {
				log.Infof("Node %s is not in status UN yet, but %s", n["address"], n["status"])
			}
		}

		return upNormalNodes
	}, 5*time.Minute, 15*time.Second).Should(Equal(nodeCount))
}

func PrintPodLogs(client client.Client, namespace, containerName string) {
	pods, err := kubernetes.ListPods(client, namespace)
	if err != nil {
		fmt.Printf("Failed to list pods: %v\n", err)
	} else {
		for _, p := range pods {
			log, err2 := p.ContainerLogs(containerName)
			if err2 != nil {
				fmt.Printf("Failed to get log for container '%s' from pod %s\n", containerName, p.Name)
			} else {
				fmt.Printf("Log for container '%s' from pod %s:\n%s\n", containerName, p.Name, log)
			}
		}
	}
}

func IsLocalCluster() bool {
	return os.Getenv("LOCAL_CLUSTER") == "true"
}

// SetSuitesParameters adds a set of common parameters
// - parameters that should be set across all testing suites
// - parameters for local testing in a minikube or other restricted environments
// This includes limited CPU and memory settings as well as disabling the Prometheus exporter when running in local cluster
func SetSuitesParameters(parameters map[string]string) {
	if IsLocalCluster() {
		parameters["NODE_MEM_MIB"] = "768"
		parameters["NODE_MEM_LIMIT_MIB"] = "1024"
		parameters["NODE_CPU_MC"] = "1000"
		parameters["NODE_CPU_LIMIT_MC"] = "1500"
		parameters["PROMETHEUS_EXPORTER_ENABLED"] = "false"
		parameters["NODE_DOCKER_IMAGE_PULL_POLICY"] = "IfNotPresent"
		parameters["BACKUP_MEDUSA_DOCKER_IMAGE_PULL_POLICY"] = "IfNotPresent"
	}
}
