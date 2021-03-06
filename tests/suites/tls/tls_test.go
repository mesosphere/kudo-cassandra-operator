package tls

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/debug"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	"github.com/kudobuilder/test-tools/pkg/tls"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/suites"
)

var (
	TestName          = "tls-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	KubectlPath       = os.Getenv("KUBECTL_PATH")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	NodeCount = 2
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

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	_ = kubernetes.CreateNamespace(Client, TestNamespace)
	_, _ = tls.CreateCertSecret("cassandra-tls").
		WithNamespace(TestNamespace).
		WithCommonName("CassandraCA").
		Do(Client)
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

var _ = Describe(TestName, func() {
	Context("Installs the operator with node-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			var err error

			parameters := map[string]string{
				"NODE_COUNT":                   strconv.Itoa(NodeCount),
				"TLS_SECRET_NAME":              "cassandra-tls",
				"TRANSPORT_ENCRYPTION_ENABLED": "true",
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

			By("Checking the container logs")
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)

			pod, err := kubernetes.GetPod(Client, podName, TestNamespace)
			Expect(err).To(BeNil())

			outputBytes, err := pod.ContainerLogs("cassandra")
			Expect(err).To(BeNil())
			Expect(string(outputBytes)).To(ContainSubstring("Starting Encrypted Messaging Service on SSL port"))

			By("Testing data read & write using CQLSH")
			output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))
		})

		It("Uninstalls the operator", func() {
			err := cassandra.Uninstall(Client, Operator)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with client-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			var err error

			parameters := map[string]string{
				"NODE_COUNT":                          strconv.Itoa(NodeCount),
				"TLS_SECRET_NAME":                     "cassandra-tls",
				"TRANSPORT_ENCRYPTION_CLIENT_ENABLED": "true",
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

			By("Checking the container logs")
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)

			pod, err := kubernetes.GetPod(Client, podName, TestNamespace)
			Expect(err).To(BeNil())

			outputBytes, err := pod.ContainerLogs("cassandra")
			Expect(err).To(BeNil())
			Expect(string(outputBytes)).To(ContainSubstring("Enabling encrypted CQL connections between client and server"))

			By("Testing data read & write using CQLSH")
			output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := cassandra.Uninstall(Client, Operator)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with node-to-node and client-to-node encryption enabled", func() {
		It("Installs the operator from a directory", func() {
			var err error

			parameters := map[string]string{
				"NODE_COUNT":                          strconv.Itoa(NodeCount),
				"TLS_SECRET_NAME":                     "cassandra-tls",
				"TRANSPORT_ENCRYPTION_ENABLED":        "true",
				"TRANSPORT_ENCRYPTION_CLIENT_ENABLED": "true",
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

			By("Checking the container logs")
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)

			pod, err := kubernetes.GetPod(Client, podName, TestNamespace)
			Expect(err).To(BeNil())

			outputBytes, err := pod.ContainerLogs("cassandra")
			Expect(err).To(BeNil())
			Expect(string(outputBytes)).To(ContainSubstring("Starting Encrypted Messaging Service on SSL port"))
			Expect(string(outputBytes)).To(ContainSubstring("Enabling encrypted CQL connections between client and server"))

			By("Testing data read & write using CQLSH")
			output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))
		})
		It("Uninstalls the operator", func() {
			err := cassandra.Uninstall(Client, Operator)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})

	Context("Installs the operator with encrypted remote JMX", func() {
		It("Installs the operator from a directory", func() {
			var err error

			parameters := map[string]string{
				"NODE_COUNT":      strconv.Itoa(NodeCount),
				"TLS_SECRET_NAME": "cassandra-tls",
				"JMX_LOCAL_ONLY":  "false",
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

			By("Checking nodetool access from an utlity pod")
			podName := fmt.Sprintf("%s-%s-%d", TestInstance, "node", 0)
			nodetool := cassandra.NewNodeTool(Client, TestNamespace, TestInstance, true)
			output, _, err := nodetool.Run("-h", fmt.Sprintf("%s.%s-svc.%s.svc.cluster.local", podName, TestInstance, TestNamespace), "--ssl", "info")
			Expect(output).To(ContainSubstring("Native Transport active: true"))
			Expect(err).To(BeNil())
		})
		It("Uninstalls the operator", func() {
			err := cassandra.Uninstall(Client, Operator)
			Expect(err).To(BeNil())
			// TODO(mpereira) Assert that it isn't running.
		})
	})
})
