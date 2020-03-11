package fault_tolerance

import (
	"fmt"
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

var _ = Describe("Fault tolerance tests", func() {
	const (
		testNamespace = "fault-tolerance-test-namespace"
	)

	var (
		client     testclient.Client
		operator   kudo.Operator
		parameters map[string]string
	)

	AfterEach(func() {
		err := operator.Uninstall()
		Expect(err).NotTo(HaveOccurred())

		err = kubernetes.DeleteNamespace(client, testNamespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when configured with the 'GossipingPropertyFileSnitch' snitch", func() {
		It("should set up the datacenter and rack properties", func() {

			// This is the nodeSelection label name for datacenters (if empty, the label will not be set)
			datacenterLabel := "failure-domain.beta.kubernetes.io/zone"

			// This is the nodeSelection label name for racks (if empty, the label will not be set)
			rackLabel := ""

			topology := cassandra.NodeTopology{
				{
					Datacenter: "us-west-2a",
					Rack:       "rac1",
					Nodes:      2,
				},
				{
					Datacenter: "us-west-2b",
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
				"PROMETHEUS_EXPORTER_ENABLED":          "false",
				"NODE_READINESS_PROBE_INITIAL_DELAY_S": "10",
			}

			By("Waiting for the operator to deploy")

			client, err = testclient.NewForConfig(kubeConfigPath)
			Expect(err).NotTo(HaveOccurred())

			err = kubernetes.CreateNamespace(client, testNamespace)
			Expect(err).NotTo(HaveOccurred())

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

			By("Checking that node status reports correct data center and rack")

			var nodesInDC1 int
			var nodesInDC2 int

			for _, node := range nodes {
				switch node["datacenter"] {
				case topology[0].Datacenter:
					nodesInDC1 += 1
					Expect(node["rack"]).To(Equal(topology[0].Rack))
				case topology[1].Datacenter:
					nodesInDC2 += 1
					Expect(node["rack"]).To(Equal(topology[1].Rack))
				default:
					Fail(fmt.Sprintf("unknown datacenter: %s", node["datacenter"]))
				}

			}

			Expect(nodesInDC1).To(Equal(topology[0].Nodes))
			Expect(nodesInDC2).To(Equal(topology[1].Nodes))
		})

		// TODO: test pod anti-affinity

		// TODO: test node selection
	})
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("fault-tolerance-test-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Fault tolerance tests", []Reporter{junitReporter})
}
