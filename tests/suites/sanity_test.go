package suites

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
	"github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

var (
	TestName          = "sanity-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	KubectlPath       = os.Getenv("KUBECTL_PATH")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")
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

var _ = Describe(TestName, func() {
	It("Installs the latest operator from the package registry", func() {
		// TODO(mpereira) Assert that it isn't running.
		err := kudo.InstallOperator(
			OperatorName, TestNamespace, TestInstance, []string{},
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

	It("Upgrades the running operator instance from a directory", func() {
		// TODO(mpereira) Assert that it is running.
		err := kudo.UpgradeOperator(
			OperatorDirectory, TestNamespace, TestInstance, []string{},
		)
		if err != nil {
			Fail(
				"Failing the full suite: failed to upgrade operator instance that the " +
					"following tests depend on",
			)
		}
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Scales the instance's number of nodes", func() {
		NodeCount = NodeCount + 1
		err := kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{"NODE_COUNT": strconv.Itoa(NodeCount)},
		)
		if err != nil {
			Fail("Failing the full suite: failed to scale the number of nodes")
		}
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Updates the instance's parameters", func() {
		parameter := "disk_failure_policy"
		initialValue := "stop"
		desiredValue := "ignore"

		configuration, err := cassandra.ClusterConfiguration(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{strings.ToUpper(parameter): desiredValue},
		)
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Configures Cassandra properties through custom properties", func() {
		parameter := "otc_backlog_expiration_interval_ms"
		initialValue := "200"
		desiredValue := "300"
		desiredEncodedProperties := base64.StdEncoding.EncodeToString(
			[]byte(parameter + ": " + desiredValue),
		)

		configuration, err := cassandra.ClusterConfiguration(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{"CUSTOM_CASSANDRA_YAML_BASE64": desiredEncodedProperties},
		)
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Configures Cassandra JVM options through custom options", func() {
		parameter := "-XX:CMSWaitDuration"
		initialValue := "10000"
		desiredValue := "11000"
		desiredEncodedProperties := base64.StdEncoding.EncodeToString(
			[]byte(parameter + "=" + desiredValue),
		)

		configuration, err := cassandra.NodeJvmOptions(
			TestNamespace, TestInstance,
		)

		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{"CUSTOM_JVM_OPTIONS_BASE64": desiredEncodedProperties},
		)
		Expect(err).To(BeNil())

		configuration, err = cassandra.NodeJvmOptions(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)
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
