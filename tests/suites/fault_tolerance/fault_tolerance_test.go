package fault_tolerance

import (
	"fmt"
	"github.com/thoas/go-funk"
	"os"
	"testing"
	"time"

	testclient "github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	kubeConfigPath    = os.Getenv("KUBECONFIG")
	operatorName      = os.Getenv("OPERATOR_NAME")
	operatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	instanceName = fmt.Sprintf("%s-instance", operatorName)
)

const createSchemaCQLTemplate = "CREATE SCHEMA schema1 WITH replication = { 'class' : 'NetworkTopologyStrategy', %s };"

const testCQLScript = "USE schema1;" +
	"CREATE TABLE users (user_id varchar PRIMARY KEY,first varchar,last varchar,age int);" +
	"INSERT INTO users (user_id, first, last, age) VALUES ('jsmith', 'John', 'Smith', 42);" +
	"SELECT * FROM users;"

const testCQLOutputScript = "USE schema1;" +
	"SELECT * FROM users;"

const testCQLScriptOutput = `
 user_id | age | first | last
---------+-----+-------+-------
  jsmith |  42 |  John | Smith

(1 rows)`

func buildDatacenterReplicationString(topology cassandra.NodeTopology, maxReplica int) string {
	result := ""
	for _, node := range topology {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("'%s': %d", node.Datacenter, funk.MinInt([]int{maxReplica, node.Nodes}))
	}
	return result
}

var _ = Describe("Fault tolerance tests", func() {
	const (
		testNamespace = "fault-tolerance-test-namespace"
	)

	var (
		client     testclient.Client
		operator   kudo.Operator
		parameters map[string]string
	)

	//AfterEach(func() {
	//	err := operator.Uninstall()
	//	Expect(err).NotTo(HaveOccurred())
	//
	//	err = kubernetes.DeleteNamespace(client, testNamespace)
	//	Expect(err).NotTo(HaveOccurred())
	//})

	Context("when configured with the 'GossipingPropertyFileSnitch' snitch", func() {
		It("should set up the datacenter and rack properties", func() {
			var err error

			client, err = testclient.NewForConfig(kubeConfigPath)
			Expect(err).NotTo(HaveOccurred())

			err = kubernetes.CreateNamespace(client, testNamespace)
			Expect(err).NotTo(HaveOccurred())

			// This is the nodeSelection label name for datacenters (if empty, the label will not be set)
			datacenterLabel := "failure-domain.beta.kubernetes.io/zone"

			// This is the nodeSelection label name for racks (if empty, the label will not be set)
			rackLabel := ""

			By("Installing the operator with a topology")
			//topology := cassandra.NodeTopology{
			//	{
			//		Datacenter: "us-west-2a",
			//		Rack:       "rac1",
			//		Nodes:      2,
			//	},
			//	{
			//		Datacenter: "us-west-2b",
			//		Rack:       "rac1",
			//		Nodes:      2,
			//	},
			//}
			topology := cassandra.NodeTopology{
				{
					Datacenter: "us-west-2a",
					Rack:       "rac1",
					Nodes:      3,
				},
				{
					Datacenter: "us-west-2b",
					Rack:       "rac1",
					Nodes:      2,
				},
				{
					Datacenter: "us-west-2c",
					Rack:       "rac1",
					Nodes:      2,
				},
			}
			topologyYaml, err := topology.ToYAML()
			Expect(err).NotTo(HaveOccurred())

			parameters = map[string]string{
				"NODE_COUNT":                           "1", // NODE_TOPOLOGY should override this value
				"ENDPOINT_SNITCH":                      "GossipingPropertyFileSnitch",
				"NODE_TOPOLOGY":                        topologyYaml,
				"DATACENTER_LABEL":                     datacenterLabel,
				"RACK_LABEL":                           rackLabel,
				"NODE_ANTI_AFFINITY":                   "true",
				"NODE_READINESS_PROBE_INITIAL_DELAY_S": "10",
			}

			By("Waiting for the operator to deploy")
			operator, err = kudo.InstallOperator(operatorDirectory).
				WithNamespace(testNamespace).
				WithInstance(instanceName).
				WithParameters(parameters).
				Do(client)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			By("Ensuring that all nodes are up")

			nodes, err := cassandra.Nodes(client, operator.Instance)
			Expect(err).NotTo(HaveOccurred())

			var totalNodes = 0
			for _, dc := range topology {
				totalNodes += dc.Nodes
			}
			Expect(nodes).To(HaveLen(totalNodes))

			By("Writing data to the cluster")
			replicationString := buildDatacenterReplicationString(topology, 2)
			createSchemaCQL := fmt.Sprintf(createSchemaCQLTemplate, replicationString)

			output, err := cassandra.Cqlsh(client, operator.Instance, createSchemaCQL+testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))

			By("Checking that node status reports correct data center")
			dcCounts := collectDataCenterCounts(nodes)
			for _, dc := range topology {
				Expect(dcCounts[dc.Datacenter]).To(Equal(dc.Nodes))
			}

			By("Updating the topology")
			topology = cassandra.NodeTopology{
				{
					Datacenter: "us-west-2a",
					Rack:       "rac1",
					Nodes:      3,
				},
				{
					Datacenter: "us-west-2b",
					Rack:       "rac1",
					Nodes:      2,
				},
				{
					Datacenter: "us-west-2c",
					Rack:       "rac1",
					Nodes:      3,
				},
			}
			topologyYaml, err = topology.ToYAML()
			parameters = map[string]string{
				"NODE_TOPOLOGY": topologyYaml,
			}
			err = operator.Instance.UpdateParameters(parameters)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanInProgress("deploy", kudo.WaitTimeout(time.Minute*2))
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			By("Ensuring that all nodes are up")

			nodes, err = cassandra.Nodes(client, operator.Instance)
			Expect(err).NotTo(HaveOccurred())

			dcCounts = collectDataCenterCounts(nodes)
			for _, dc := range topology {
				Expect(dcCounts[dc.Datacenter]).To(Equal(dc.Nodes))
			}

			By("Reading data from the cluster")
			output, err = cassandra.Cqlsh(client, operator.Instance, testCQLOutputScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))

			By("Checking that nodes are deployed on different ndoes")
			podList, err := kubernetes.ListPods(client, testNamespace)
			Expect(err).To(BeNil())

			usedIPs := map[string]bool{}
			for _, pod := range podList {
				ip := pod.Status.HostIP
				_, ok := usedIPs[ip]
				Expect(ok).To(BeFalse(), "HostIP has been reused, anti-affinity should prevent that")
			}

		})

		// TODO: test node selection
	})
})

func collectDataCenterCounts(nodes []map[string]string) map[string]int {
	result := map[string]int{}
	for _, node := range nodes {
		val, ok := result[node["datacenter"]]
		if !ok {
			result[node["datacenter"]] = 1
		} else {
			result[node["datacenter"]] = val + 1
		}
	}
	return result
}

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("fault-tolerance-test-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Fault tolerance tests", []Reporter{junitReporter})
}
