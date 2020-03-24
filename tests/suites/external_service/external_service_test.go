package external_service

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	"github.com/onsi/ginkgo/reporters"
	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/suites"
)

var (
	TestName          = "ext-service-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	NodeCount = 1
	Client    = client.Client{}
	Operator  = kudo.Operator{}
)

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	_ = kubernetes.CreateNamespace(Client, TestNamespace)
})

var _ = AfterSuite(func() {
	_ = Operator.Uninstall()
	_ = kubernetes.DeleteNamespace(Client, TestNamespace)
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}

func assertNumberOfCassandraNodes(nodeCount int) {
	nodes, err := cassandra.Nodes(Client, Operator.Instance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = Describe("external service", func() {

	It("Installs the operator from the current directory", func() {
		var err error

		parameters := map[string]string{
			"NODE_COUNT": strconv.Itoa(NodeCount),
		}
		suites.SetLocalClusterParameters(parameters)

		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(parameters).
			Do(Client)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		By("Allowing external access to the cassandra cluster")
		nativeTransportPort := 9043
		parameters = map[string]string{
			"EXTERNAL_SERVICE":               "true",
			"EXTERNAL_NATIVE_TRANSPORT":      "true",
			"EXTERNAL_NATIVE_TRANSPORT_PORT": strconv.Itoa(nativeTransportPort),
		}
		suites.SetLocalClusterParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		log.Infof("Verify that external service is started and has 1 open port")
		svc, err := kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(1))
		Expect(svc.Spec.Ports[0].Name).To(Equal("native-transport"))
		Expect(svc.Spec.Ports[0].Port).To(Equal(int32(nativeTransportPort)))

		By("Opening a second port if rpc is enabled")
		rpcPort := 9161
		parameters = map[string]string{
			"START_RPC":         "true",
			"EXTERNAL_RPC":      "true",
			"EXTERNAL_RPC_PORT": strconv.Itoa(rpcPort),
		}
		suites.SetLocalClusterParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)
		Expect(err).To(BeNil())

		//err = Operator.Instance.WaitForPlanInProgress("deploy", kudo.WaitTimeout(time.Second*90))
		//Expect(err).To(BeNil())
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		log.Infof("Verify that external service is started and has 2 open ports")
		svc, err = kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(2))
		Expect(svc.Spec.Ports[1].Name).To(Equal("rpc"))
		Expect(svc.Spec.Ports[1].Port).To(Equal(int32(rpcPort)))

		By("Disabling the external service again")
		parameters = map[string]string{
			"START_RPC":                 "false",
			"EXTERNAL_SERVICE":          "false",
			"EXTERNAL_RPC":              "false",
			"EXTERNAL_NATIVE_TRANSPORT": "false",
		}
		suites.SetLocalClusterParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)
		Expect(err).To(BeNil())

		//err = Operator.Instance.WaitForPlanInProgress("deploy", kudo.WaitTimeout(time.Second*90))
		//Expect(err).To(BeNil())
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		svc, err = kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)

		log.Infof("Get External Service %v, %v", svc, err)

		Expect(err).To(Not(BeNil()))

	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
	})
})
