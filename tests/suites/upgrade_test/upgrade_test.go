package upgrade_test

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
	TestName          = "upgrade-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
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
	It("Upgrade the operator", func() {
		err := kudo.UpgradeOperatorFromDirectory(
			"https://infinity-artifacts.s3-us-west-2.amazonaws.com/cassandra/cassandra-0.1.1.tgz", TestNamespace, TestInstance, []string{},
		)
		// TODO(mpereira) Assert that it is running.
		if err != nil {
			Fail(
				"Failing the full suite: failed to upgrade operator instance that the " +
					"following tests depend on",
			)
		}
		Expect(err).To(BeNil())
	})
})

var _ = BeforeSuite(func() {
	k8s.Init(KubectlOptions)
	kudo.Init(KubectlOptions)
	kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	k8s.CreateNamespace(TestNamespace)
	kudo.InstallOperatorFromDirectory(
		OperatorDirectory, TestNamespace, TestInstance, []string{},
	)
})

var _ = AfterSuite(func() {
	kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	k8s.DeleteNamespace(TestNamespace)
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}
