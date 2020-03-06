package upgrade

import (
	"fmt"
	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var (
	TestName            = "sanity-test"
	TestOperatorVersion = "99.99.99-testing"
	OperatorVersion     string
	OperatorName        = os.Getenv("OPERATOR_NAME")
	TestNamespace       = fmt.Sprintf("%s-namespace", TestName)
	TestInstance        = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath      = os.Getenv("KUBECONFIG")
	OperatorDirectory   = os.Getenv("OPERATOR_DIRECTORY")

	// Node Count of 1 for Sanity test to have the tests a little bit faster
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

// This test is disabled for now, as the operator does not really support upgrading yet
//func TestService(t *testing.T) {
//	RegisterFailHandler(Fail)
//	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
//		"%s-junit.xml", TestName,
//	))
//	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
//}

func assertNumberOfCassandraNodes(nodeCount int) {
	nodes, err := cassandra.Nodes(Client, Operator.Instance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = Describe(TestName, func() {
	It("Installs the latest operator from the package registry", func() {
		var err error

		// TODO(mpereira) Assert that it isn't running.
		Operator, err = kudo.InstallOperator(OperatorName).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(map[string]string{
				"NODE_COUNT":                  strconv.Itoa(NodeCount),
				"PROMETHEUS_EXPORTER_ENABLED": "false",
			}).
			Do(Client)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())
		assertNumberOfCassandraNodes(NodeCount)

		By("Upgrading the running operator instance from a directory")

		before, _, err := cassandra.OverrideOperatorVersion(TestOperatorVersion)
		if err != nil {
			log.Errorf(
				"Error overriding operatorVersion from '%s' to '%s': %v",
				OperatorVersion, TestOperatorVersion, err,
			)
		}
		OperatorVersion = before
		Expect(err).To(BeNil())

		err = kudo.UpgradeOperator().WithOperator(OperatorDirectory).Do(&Operator)
		if err != nil {
			Fail(
				"Failing the full suite: failed to upgrade operator instance that the " +
					"following tests depend on",
			)
		}
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("upgrade")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("upgrade")
		Expect(err).To(BeNil())
		assertNumberOfCassandraNodes(NodeCount)
		//It("Uninstalls the operator", func() {
		//	err := cassandra.Uninstall(Client, Operator)
		//	Expect(err).To(BeNil())
		//	// TODO(mpereira) Assert that it isn't running.
		//})
	})

})
