package suites

import (
	"fmt"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

var (
	TestName            = "sanity-test"
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

func assertNumberOfCassandraNodes(nodeCount int) {
	nodes, err := cassandra.Nodes(TestNamespace, TestInstance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = BeforeSuite(func() {
	k8s.Init(KubectlOptions)
	kudo.Init(KubectlOptions)

	kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	k8s.CreateNamespace(TestNamespace)
})

var _ = BeforeEach(func() {
})

var _ = AfterEach(func() {
})

var _ = AfterSuite(func() {
	kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	k8s.DeleteNamespace(TestNamespace)
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
