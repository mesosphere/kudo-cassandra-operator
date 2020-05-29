package suites

import (
	"os"
	"time"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

func AssertNumberOfCassandraNodes(client client.Client, op kudo.Operator, nodeCount int) {
	Eventually(func() int {
		nodes, err := cassandra.Nodes(client, op.Instance)
		Expect(err).To(BeNil())

		return len(nodes)
	}, 5*time.Minute, 15*time.Second).Should(Equal(nodeCount))
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
