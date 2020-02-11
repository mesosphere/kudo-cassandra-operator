package tls

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	"github.com/kudobuilder/test-tools/pkg/tls"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	TestName          = "tls-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")
	// TODO(mpereira): read NodeCount from params.yaml.
	NodeCount = 3
	Client    = client.Client{}
	Operator  = kudo.Operator{}
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
	nodes, err := cassandra.Nodes(Client, Operator.Instance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = Describe(TestName, func() {
	Context("Installs the operator with node-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			var err error

			Operator, err = kudo.InstallOperator(OperatorDirectory).
				WithNamespace(TestNamespace).
				WithInstance(TestInstance).
				WithParameters(map[string]string{
					"TLS_SECRET_NAME":              "cassandra-tls",
					"TRANSPORT_ENCRYPTION_ENABLED": "true",
				}).
				Do(Client)
			Expect(err).To(BeNil())

			err = Operator.Instance.WaitForPlanInProgress("deploy")
			Expect(err).To(BeNil())

			err = Operator.Instance.WaitForPlanComplete("deploy")
			Expect(err).To(BeNil())

			assertNumberOfCassandraNodes(NodeCount)
		})
		It("Checks for the container logs", func() {
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)

			pod, err := kubernetes.GetPod(Client, podName, TestNamespace)
			Expect(err).To(BeNil())

			output, err := pod.ContainerLogs("cassandra")
			Expect(err).To(BeNil())
			Expect(string(output)).To(ContainSubstring("Starting Encrypted Messaging Service on SSL port"))
		})
		It("Tests data read & write using CQLSH", func() {
			output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := Operator.Uninstall()
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with client-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			var err error

			Operator, err = kudo.InstallOperator(OperatorDirectory).
				WithNamespace(TestNamespace).
				WithInstance(TestInstance).
				WithParameters(map[string]string{
					"TLS_SECRET_NAME":                     "cassandra-tls",
					"TRANSPORT_ENCRYPTION_CLIENT_ENABLED": "true",
				}).
				Do(Client)
			Expect(err).To(BeNil())

			err = Operator.Instance.WaitForPlanInProgress("deploy")
			Expect(err).To(BeNil())

			err = Operator.Instance.WaitForPlanComplete("deploy")
			Expect(err).To(BeNil())

			assertNumberOfCassandraNodes(NodeCount)
		})
		It("Checks for the container logs", func() {
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)

			pod, err := kubernetes.GetPod(Client, podName, TestNamespace)
			Expect(err).To(BeNil())

			output, err := pod.ContainerLogs("cassandra")
			Expect(err).To(BeNil())
			Expect(string(output)).To(ContainSubstring("Enabling encrypted CQL connections between client and server"))
		})
		It("Tests data read & write using CQLSH", func() {
			output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := Operator.Uninstall()
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with node-to-node and client-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			var err error

			Operator, err = kudo.InstallOperator(OperatorDirectory).
				WithNamespace(TestNamespace).
				WithInstance(TestInstance).
				WithParameters(map[string]string{
					"TLS_SECRET_NAME":                     "cassandra-tls",
					"TRANSPORT_ENCRYPTION_ENABLED":        "true",
					"TRANSPORT_ENCRYPTION_CLIENT_ENABLED": "true",
				}).
				Do(Client)
			Expect(err).To(BeNil())

			err = Operator.Instance.WaitForPlanInProgress("deploy")
			Expect(err).To(BeNil())

			err = Operator.Instance.WaitForPlanComplete("deploy")
			Expect(err).To(BeNil())

			assertNumberOfCassandraNodes(NodeCount)
		})
		It("Checks for the container logs", func() {
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)

			pod, err := kubernetes.GetPod(Client, podName, TestNamespace)
			Expect(err).To(BeNil())

			output, err := pod.ContainerLogs("cassandra")
			Expect(err).To(BeNil())
			Expect(string(output)).To(ContainSubstring("Starting Encrypted Messaging Service on SSL port"))
			Expect(string(output)).To(ContainSubstring("Enabling encrypted CQL connections between client and server"))
		})
		It("Tests data read & write using CQLSH", func() {
			output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := Operator.Uninstall()
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})
})

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	kubernetes.CreateNamespace(Client, TestNamespace)
	tls.CreateCertSecret("cassandra-tls").
		WithNamespace(TestNamespace).
		WithCommonName("CassandraCA").
		Do(Client)
})

var _ = AfterSuite(func() {
	Operator.Uninstall()
	kubernetes.DeleteNamespace(Client, TestNamespace)
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}
