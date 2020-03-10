package fault_tolerance

import (
	"fmt"
	"os"
	"testing"

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

	JustBeforeEach(func() {
		var err error

		client, err = testclient.NewForConfig(kubeConfigPath)
		Expect(err).NotTo(HaveOccurred())

		err = kubernetes.CreateNamespace(client, testNamespace)
		Expect(err).NotTo(HaveOccurred())

		operator, err = kudo.InstallOperator(operatorDirectory).
			WithNamespace(testNamespace).
			WithInstance(fmt.Sprintf("%s-instance", operatorName)).
			WithParameters(parameters).
			Do(client)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := operator.Uninstall()
		Expect(err).NotTo(HaveOccurred())

		err = kubernetes.DeleteNamespace(client, testNamespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when configured with the 'GossipingPropertyFileSnitch' snitch", func() {
		BeforeEach(func() {
			topology, err := cassandra.NodeTopology{
				{
					Datacenter: "dc1",
					Rack: "rac1",
					Nodes: 2,
				},
				{
					Datacenter: "dc2",
					Rack: "rac1",
					Nodes: 1,
				},
			}.ToYAML()
			Expect(err).NotTo(HaveOccurred())

			parameters = map[string]string{
				"NODE_COUNT": "1", // NODE_TOPOLOGY should override this value
				"ENDPOINT_SNITCH": "GossipingPropertyFileSnitch",
				"NODE_TOPOLOGY": topology,
				"DATACENTER_LABEL": "",
				"RACK_LABEL": "",
			}
		})

		It("should set up the datacenter and rack properties", func() {
			By("Waiting for the operator to deploy")

			err := operator.Instance.WaitForPlanComplete("deploy")
			Expect(err).NotTo(HaveOccurred())

			By("Ensuring that all nodes are up")

			nodes, err := cassandra.Nodes(client, operator.Instance)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodes).To(HaveLen(3))

			By("Checking that node status reports correct datacenter and rack")

			var nodesInDC1 int
			var nodesInDC2 int

			for _, node := range nodes {
				switch node["datacenter"] {
				case "dc1":
					nodesInDC1 += 1
				case "dc2":
					nodesInDC2 += 1
				default:
					Fail("unknown datacenter")
				}

				Expect(node["rack"]).To(Equal("rac1"))
			}

			Expect(nodesInDC1).To(Equal(2))
			Expect(nodesInDC2).To(Equal(1))
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
