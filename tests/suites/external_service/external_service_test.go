package external_service

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	"github.com/onsi/ginkgo/reporters"
	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	TestName            = "ext-service-test"
	TestOperatorVersion = "99.99.99-testing"
	OperatorVersion     string
	OperatorName        = os.Getenv("OPERATOR_NAME")
	TestNamespace       = fmt.Sprintf("%s-namespace", TestName)
	TestInstance        = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath      = os.Getenv("KUBECONFIG")
	KubectlPath         = os.Getenv("KUBECTL_PATH")
	OperatorDirectory   = os.Getenv("OPERATOR_DIRECTORY")

	NodeCount = 1
	Client    = client.Client{}
	Operator  = kudo.Operator{}
)

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	kubernetes.CreateNamespace(Client, TestNamespace)
})

var _ = AfterSuite(func() {
	Operator.Uninstall()
	kubernetes.DeleteNamespace(Client, TestNamespace)
})

//var _ = BeforeSuite(func() {
//	k8s.Init(KubectlOptions)
//	kudo.Init(KubectlOptions)
//	_ = kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
//	_ = k8s.CreateNamespace(TestNamespace)
//})
//
//var _ = AfterSuite(func() {
//	_ = kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
//	_ = k8s.DeleteNamespace(TestNamespace)
//	if OperatorVersion != "" {
//		_, _, err := kudo.OverrideOperatorVersion(OperatorVersion)
//		if err != nil {
//			log.Errorf(
//				"Error reverting operatorVersion from '%s' to '%s': %v",
//				TestOperatorVersion, OperatorVersion, err,
//			)
//		}
//	}
//})

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

		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(map[string]string{
				"NODE_COUNT": strconv.Itoa(NodeCount),
			}).
			Do(Client)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy", kudo.WaitTimeout(time.Second*90))
		Expect(err).To(BeNil())
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Allows external access to the cassandra cluster", func() {
		nativeTransportPort := 9043
		err := Operator.Instance.UpdateParameters(map[string]string{
			"EXTERNAL_NATIVE_TRANSPORT":      "true",
			"EXTERNAL_NATIVE_TRANSPORT_PORT": strconv.Itoa(nativeTransportPort),
		})
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		log.Infof("Verify that external service is started and has 1 open port")
		svc, err := kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(1))
		Expect(svc.Spec.Ports[0].Name).To(Equal("native-transport"))
		Expect(svc.Spec.Ports[0].Port).To(Equal(int32(nativeTransportPort)))
	})

	It("Opens a second port if rpc is enabled", func() {
		rpcPort := 9161
		err := Operator.Instance.UpdateParameters(map[string]string{
			"START_RPC":         "true",
			"EXTERNAL_RPC":      "true",
			"EXTERNAL_RPC_PORT": strconv.Itoa(rpcPort),
		})
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		log.Infof("Verify that external service is started and has 2 open ports")
		svc, err := kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(2))
		Expect(svc.Spec.Ports[1].Name).To(Equal("rpc"))
		Expect(svc.Spec.Ports[1].Port).To(Equal(int32(rpcPort)))
	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
	})
})
