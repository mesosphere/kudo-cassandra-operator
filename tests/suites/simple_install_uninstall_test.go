package suites

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	// log "github.com/sirupsen/logrus"

	k8s "github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	kubectl "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
	kudo "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

var (
	TestName          = "simple-install-uninstall-test"
	OperatorName      = "cassandra"
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	KubectlPath       = os.Getenv("KUBECTL_PATH")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")
	KubectlOptions    = kubectl.NewKubectlOptions(
		KubectlPath,
		KubeConfigPath,
		TestNamespace,
		"",
	)
)

var _ = Describe(TestName, func() {
	It("Installs the operator from a directory", func() {
		// TODO(mpereira) Assert that it isn't running.
		err := kudo.InstallOperatorFromDirectory(
			OperatorDirectory, TestNamespace, TestInstance, []string{},
		)
		Expect(err).To(BeNil())
		// TODO(mpereira) Assert that it is running.
	})
	It("Uninstalls the operator", func() {
		err := kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
		Expect(err).To(BeNil())
		// TODO(mpereira) Assert that it isn't running.
	})
})

var _ = BeforeSuite(func() {
	k8s.Init(KubectlOptions)
	kudo.Init(KubectlOptions)
	kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	k8s.CreateNamespace(TestNamespace)
})

var _ = AfterSuite(func() {
	k8s.DeleteNamespace(TestNamespace)
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}
