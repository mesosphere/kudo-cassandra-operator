package tls

import (
	"fmt"
	"os"
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
	TestName          = "tls-test"
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

const testCQLScript = "CREATE SCHEMA schema1 WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };" +
	"USE schema1;" +
	"CREATE TABLE users (user_id varchar PRIMARY KEY,first varchar,last varchar,age int);" +
	"INSERT INTO users (user_id, first, last, age) VALUES ('jsmith', 'John', 'Smith', 42);" +
	"SELECT * FROM users;"

const testCQLScriptOutput = `
 user_id | age | first | last
---------+-----+-------+-------
  jsmith |  42 |  John | Smith

(1 rows)`

func assertNumberOfCassandraNodes(nodeCount int) {
	nodes, err := cassandra.Nodes(TestNamespace, TestInstance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = Describe(TestName, func() {
	Context("Installs the operator with node-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			err := kudo.InstallOperator(
				OperatorDirectory, TestNamespace, TestInstance, []string{
					"TLS_SECRET_NAME=cassandra-tls",
					"TRANSPORT_ENCRYPTION_ENABLED=true",
				},
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
		It("Checks for the container logs", func() {
			output, _ := k8s.GetPodContainerLogs(
				TestNamespace,
				fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0),
				"cassandra",
			)
			Expect(output).To(ContainSubstring("Starting Encrypted Messaging Service on SSL port"))
		})
		It("Tests data read & write using CQLSH", func() {
			output, err := cassandra.Cqlsh(
				TestNamespace,
				TestInstance,
				testCQLScript,
			)
			Expect(err).To(BeNil())
			Expect(output.String()).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with client-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			err := kudo.InstallOperator(
				OperatorDirectory, TestNamespace, TestInstance, []string{
					"TLS_SECRET_NAME=cassandra-tls",
					"TRANSPORT_ENCRYPTION_CLIENT_ENABLED=true",
				},
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
		It("Checks for the container logs", func() {
			output, _ := k8s.GetPodContainerLogs(
				TestNamespace,
				fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0),
				"cassandra",
			)
			Expect(output).To(ContainSubstring("Enabling encrypted CQL connections between client and server"))
		})
		It("Tests data read & write using CQLSH", func() {
			output, err := cassandra.Cqlsh(
				TestNamespace,
				TestInstance,
				testCQLScript,
			)
			Expect(err).To(BeNil())
			Expect(output.String()).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with node-to-node and client-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			err := kudo.InstallOperator(
				OperatorDirectory, TestNamespace, TestInstance, []string{
					"TLS_SECRET_NAME=cassandra-tls",
					"TRANSPORT_ENCRYPTION_ENABLED=true",
					"TRANSPORT_ENCRYPTION_CLIENT_ENABLED=true",
				},
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
		It("Checks for the container logs", func() {
			output, _ := k8s.GetPodContainerLogs(
				TestNamespace,
				fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0),
				"cassandra",
			)
			Expect(output).To(ContainSubstring("Starting Encrypted Messaging Service on SSL port"))
			Expect(output).To(ContainSubstring("Enabling encrypted CQL connections between client and server"))
		})
		It("Tests data read & write using CQLSH", func() {
			output, err := cassandra.Cqlsh(
				TestNamespace,
				TestInstance,
				testCQLScript,
			)
			Expect(err).To(BeNil())
			Expect(output.String()).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})
})

var _ = BeforeSuite(func() {
	k8s.Init(KubectlOptions)
	kudo.Init(KubectlOptions)
	kudo.UninstallOperator(OperatorName, TestNamespace, TestInstance)
	k8s.CreateNamespace(TestNamespace)
	k8s.CreateTLSCertSecret(TestNamespace, "cassandra-tls", "CassandraCA")
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
