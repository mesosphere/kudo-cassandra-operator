package externalservice

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/debug"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/suites"
)

var (
	TestName          = "ext-service-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	KubectlPath       = os.Getenv("KUBECTL_PATH")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	NodeCount = 1
	Client    = client.Client{}
	Operator  = kudo.Operator{}
)

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	_ = kubernetes.CreateNamespace(Client, TestNamespace)
})

var _ = AfterEach(func() {
	debug.CollectArtifacts(Client, afero.NewOsFs(), GinkgoWriter, TestNamespace, KubectlPath)
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

var _ = Describe("external service", func() {

	It("Installs the operator from the current directory", func() {
		var err error

		parameters := map[string]string{
			"NODE_COUNT": strconv.Itoa(NodeCount),
		}
		suites.SetSuitesParameters(parameters)

		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(parameters).
			Do(Client)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		By("Allowing external access to the cassandra cluster")
		nativeTransportPort := 9043
		parameters = map[string]string{
			"EXTERNAL_SERVICE":               "true",
			"EXTERNAL_NATIVE_TRANSPORT":      "true",
			"EXTERNAL_NATIVE_TRANSPORT_PORT": strconv.Itoa(nativeTransportPort),
		}
		suites.SetSuitesParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		log.Infof("Verify that external service is started and has 1 open port")
		Eventually(func() bool {
			svc, _ := kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
			log.Infof("External Service: %v, Error: %v", svc, err)
			return err != nil && len(svc.Spec.Ports) == 1 && svc.Spec.Ports[0].Name == "native-transport" && svc.Spec.Ports[0].Port == int32(nativeTransportPort)
		}, 2*time.Minute, 10*time.Second).Should(BeTrue())

		By("Opening a second port if rpc is enabled")
		rpcPort := 9161
		parameters = map[string]string{
			"START_RPC":         "true",
			"EXTERNAL_RPC":      "true",
			"EXTERNAL_RPC_PORT": strconv.Itoa(rpcPort),
		}
		suites.SetSuitesParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		log.Infof("Verify that external service is started and has 2 open ports")
		Eventually(func() bool {
			svc, _ := kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
			log.Infof("External Service: %v, Error: %v", svc, err)
			return err != nil && len(svc.Spec.Ports) == 2 && svc.Spec.Ports[1].Name == "rpc" && svc.Spec.Ports[1].Port == int32(rpcPort)
		}, 2*time.Minute, 10*time.Second).Should(BeTrue())

		By("Disabling the external service again")
		parameters = map[string]string{
			"START_RPC":                 "false",
			"EXTERNAL_SERVICE":          "false",
			"EXTERNAL_RPC":              "false",
			"EXTERNAL_NATIVE_TRANSPORT": "false",
		}
		suites.SetSuitesParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)
		Expect(err).To(BeNil())

		//err = Operator.Instance.WaitForPlanInProgress("deploy", kudo.WaitTimeout(time.Second*90))
		//Expect(err).To(BeNil())
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		Eventually(func() error {
			svc, err := kubernetes.GetService(Client, fmt.Sprintf("%s-svc-external", TestInstance), TestNamespace)
			log.Infof("External Service: %v, Error: %v", svc, err)
			return err
		}, 2*time.Minute, 10*time.Second).Should(Not(BeNil()))

		Expect(err).To(Not(BeNil()))
	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
	})
})
