package external_service

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	TestName            = "backup-restore-test"
	TestOperatorVersion = "99.99.99-testing"
	OperatorVersion     string
	OperatorName        = os.Getenv("OPERATOR_NAME")
	TestNamespace       = fmt.Sprintf("%s", TestName)
	TestInstance        = fmt.Sprintf("%s", OperatorName)
	RestoreInstance     = fmt.Sprintf("%s-restore", OperatorName)
	KubeConfigPath      = os.Getenv("KUBECONFIG")
	KubectlPath         = os.Getenv("KUBECTL_PATH")
	OperatorDirectory   = os.Getenv("OPERATOR_DIRECTORY")

	NodeCount = 2
	Client    = client.Client{}
	Operator  = kudo.Operator{}

	BackupPrefix = "test2"
	BackupName   = "first"
)

const createSchema = "CREATE SCHEMA schema1 WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"
const useSchma = "USE schema1;"
const createTable = "CREATE TABLE users (user_id varchar PRIMARY KEY,first varchar,last varchar,age int);"
const insertData = "INSERT INTO users (user_id, first, last, age) VALUES ('jsmith', 'John', 'Smith', 42);"
const selectData = "SELECT * FROM users;"

const testCQLScript = createSchema + useSchma + createTable + insertData + selectData

const confirmCQLScript = useSchma + selectData

const testCQLScriptOutput = `
 user_id | age | first | last
---------+-----+-------+-------
  jsmith |  42 |  John | Smith

(1 rows)`

var _ = BeforeSuite(func() {
	buildNumber := os.Getenv("BUILD_NUMBER")
	tcProject := os.Getenv("TEAMCITY_PROJECT_NAME")
	if buildNumber != "" && tcProject != "" {
		BackupPrefix = "TC-" + tcProject + "-" + buildNumber
	}
	fmt.Printf("Using backup prefix %s\n", BackupPrefix)

	Client, _ = client.NewForConfig(KubeConfigPath)
	if err := kubernetes.CreateNamespace(Client, TestNamespace); err != nil {
		fmt.Printf("Failed to create namespace: %v\n", err)
	}
})

//var _ = AfterSuite(func() {
//	if err := Operator.Uninstall(); err != nil {
//		fmt.Printf("Failed to uninstall operator: %v\n", err)
//	}
//	if err := kubernetes.DeleteNamespace(Client, TestNamespace); err != nil {
//		fmt.Printf("Failed to delete namespace: %v\n", err)
//	}
//
//	if err := aws.DeleteFolderInS3(BackupPrefix); err != nil {
//		fmt.Printf("Error while cleaning up S3 bucket: %v\n", err)
//	}
//})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}

func assertNumberOfCassandraNodes(nodeCount int) {
	Eventually(func() int {
		nodes, err := cassandra.Nodes(Client, Operator.Instance)
		Expect(err).To(BeNil())
		return len(nodes)
	}, "60s", "2s").Should(Equal(nodeCount))
}

func createAwsCredentials() string {
	awsSecretName := "aws-credentials"

	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSecurityToken := os.Getenv("AWS_SECURITY_TOKEN")

	if awsAccessKey == "" || awsSecretKey == "" {
		Fail("No AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY defined. These are required for backup to AWS S3")
	}

	awsCredentials := make(map[string]string, 2)
	awsCredentials["access-key"] = awsAccessKey
	awsCredentials["secret-key"] = awsSecretKey
	if awsSecurityToken != "" {
		awsCredentials["security-token"] = awsSecurityToken
	}

	By("Creating a aws-credentials secret")
	_, _ = kubernetes.CreateSecret(awsSecretName).
		WithNamespace(TestNamespace).
		WithStringData(awsCredentials).Do(Client)

	return awsSecretName
}

var _ = Describe("backup and restore", func() {

	It("Installs the operator from the current directory", func() {
		var err error

		awsSecretName := createAwsCredentials()

		By("Installing the operator from current directory")
		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(map[string]string{
				"NODE_COUNT":                    strconv.Itoa(NodeCount),
				"JMX_LOCAL_ONLY":                "false",
				"PROMETHEUS_EXPORTER_ENABLED":   "false",
				"NODE_MEM_MIB":                  "768",
				"NODE_MEM_LIMIT_MIB":            "1024",
				"NODE_CPU_MC":                   "1000",
				"NODE_CPU_LIMIT_MC":             "1500",
				"BACKUP_RESTORE_ENABLED":        "true",
				"BACKUP_AWS_CREDENTIALS_SECRET": awsSecretName,
				"BACKUP_PREFIX":                 BackupPrefix,
				"BACKUP_NAME":                   BackupName,
			}).
			Do(Client)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		By("Writing Data to the cassandra cluster")
		output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput))

		By("Running the backup plan")
		err = Operator.Instance.UpdateParameters(map[string]string{
			"BACKUP_TRIGGER": "2",
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("backup")
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("backup")
		Expect(err).To(BeNil())

		By("Uninstalling the operator instance")
		err = cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
		Eventually(func() int {
			pods, _ := kubernetes.ListPods(Client, TestNamespace)
			fmt.Printf("Polling pods: %v\n", len(pods))
			return len(pods)
		}, "300s", "10s").Should(Equal(0))

		By("Restoring the backup into a new instance")

		By("Installing the operator from current directory")
		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(RestoreInstance).
			WithParameters(map[string]string{
				"NODE_COUNT":                    strconv.Itoa(NodeCount),
				"JMX_LOCAL_ONLY":                "false",
				"PROMETHEUS_EXPORTER_ENABLED":   "false",
				"NODE_MEM_MIB":                  "768",
				"NODE_MEM_LIMIT_MIB":            "1024",
				"NODE_CPU_MC":                   "1000",
				"NODE_CPU_LIMIT_MC":             "1500",
				"BACKUP_RESTORE_ENABLED":        "true",
				"BACKUP_AWS_CREDENTIALS_SECRET": awsSecretName,
				"BACKUP_PREFIX":                 BackupPrefix,
				"BACKUP_NAME":                   BackupName,
				"RESTORE_FLAG":                  "true",
				"RESTORE_OLD_NAMESPACE":         TestNamespace,
				"RESTORE_OLD_NAME":              TestInstance,
			}).
			Do(Client)

		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		By("Reading Data from the cassandra cluster")
		output, err = cassandra.Cqlsh(Client, Operator.Instance, confirmCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput))

	})

	//It("Uninstalls the operator", func() {
	//	err := cassandra.Uninstall(Client, Operator)
	//	Expect(err).To(BeNil())
	//})
})
