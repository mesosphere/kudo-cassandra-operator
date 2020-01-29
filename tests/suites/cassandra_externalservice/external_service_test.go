package cassandra_externalservice

import (
	"fmt"
	"os"
	"testing"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
	"github.com/onsi/ginkgo/reporters"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
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
	// TODO(mpereira): read NodeCount from params.yaml.
	NodeCount      = 3
	KubectlOptions = kubectl.NewKubectlOptions(
		KubectlPath,
		KubeConfigPath,
		TestNamespace,
		"",
	)
)

var _ = BeforeSuite(func() {
	k8s.Init(KubectlOptions)
	kudo.Init(KubectlOptions)
	_ = kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	_ = k8s.CreateNamespace(TestNamespace)
})

var _ = AfterSuite(func() {
	_ = kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	_ = k8s.DeleteNamespace(TestNamespace)
	if OperatorVersion != "" {
		_, _, err := kudo.OverrideOperatorVersion(OperatorVersion)
		if err != nil {
			log.Errorf(
				"Error reverting operatorVersion from '%s' to '%s': %v",
				TestOperatorVersion, OperatorVersion, err,
			)
		}
	}
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}

func assertNumberOfCassandraNodes(nodeCount int) {
	nodes, err := cassandra.Nodes(TestNamespace, TestInstance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = Describe("external service", func() {
	NodeCount = 3

	It("Installs the operator from the current directory", func() {
		err := kudo.InstallOperator(
			OperatorDirectory, TestNamespace, TestInstance, []string{}, true,
		)
		if err != nil {
			Fail(
				"Failing the full suite: failed to install operator instance that the " +
					"following tests depend on",
			)
		}
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Allows external access to the cassandra cluster", func() {
		err := kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{"EXTERNAL_NATIVE_TRANSPORT": "true"},
			false,
		)
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		svc, err := k8s.GetService(TestNamespace, fmt.Sprintf("%s-svc-external", TestInstance))
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(1))
	})

	It("Opens a second port if rpc is enabled", func() {
		err := kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{
				"START_RPC":    "true",
				"EXTERNAL_RPC": "true",
			},
			true,
		)
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		svc, err := k8s.GetService(TestNamespace, fmt.Sprintf("%s-svc-external", TestInstance))
		Expect(err).To(BeNil())
		Expect(len(svc.Spec.Ports)).To(Equal(2))
	})

	It("Uninstalls the operator", func() {
		err := kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
		Expect(err).To(BeNil())
	})
})
