package external_service

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kudobuilder/test-tools/pkg/debug"
	"github.com/spf13/afero"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	"github.com/kudobuilder/test-tools/pkg/tls"
	"github.com/mesosphere/kudo-cassandra-operator/tests/aws"
	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/suites"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

// To run this test locally you need to have the AWS credentials in the env:
// maws
// source ../scripts/export_maws.sh
// ./run.sh backup_restore
//
// Additionally, you can export LOCAL_CLUSTER=true to only use very limited resources
//
// When running these tests locall, it helps to pre-load the required images with kind load docker-image <image-name>

var (
	TestName          = "backup-restore-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s", TestName)
	TestInstance      = fmt.Sprintf("%s", OperatorName)
	RestoreInstance   = fmt.Sprintf("%s-restore", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	KubectlPath       = os.Getenv("KUBECTL_PATH")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	NodeCount = 2
	Client    = client.Client{}
	Operator  = kudo.Operator{}

	BackupBucket    = "kudo-cassandra-backup-test"
	BackupPrefix    = uuid.New().String()
	BackupPrefixTls = uuid.New().String()
	BackupName      = "first"

	AwsSecretName string

	Secret     *kubernetes.Secret
	TlsSecret  *kubernetes.Secret
	AuthSecret *kubernetes.Secret
)

const createSchema = "CREATE SCHEMA schema1 WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"
const useSchema = "USE schema1;"
const createTable = "CREATE TABLE users (user_id varchar PRIMARY KEY,first varchar,last varchar,age int);"
const insertData = "INSERT INTO users (user_id, first, last, age) VALUES ('jsmith', 'John', 'Smith', 42);"
const insertData2 = "INSERT INTO users (user_id, first, last, age) VALUES ('jdoe', 'Jane', 'Doe', 23);"
const selectData = "SELECT * FROM users;"

const testCQLScript = createSchema + useSchema + createTable + insertData + selectData

const additionalDataCQLScript = useSchema + insertData2 + selectData

const confirmCQLScript = useSchema + selectData

const testCQLScriptOutput = `
 user_id | age | first | last
---------+-----+-------+-------
  jsmith |  42 |  John | Smith

(1 rows)`

const testCQLScriptOutput2 = `
 user_id | age | first | last
---------+-----+-------+-------
    jdoe |  23 |  Jane |   Doe
  jsmith |  42 |  John | Smith

(2 rows)`

var _ = BeforeSuite(func() {
	fmt.Printf("Using backup prefix %s\n", BackupPrefix)

	Client, _ = client.NewForConfig(KubeConfigPath)
	if err := kubernetes.CreateNamespace(Client, TestNamespace); err != nil {
		fmt.Printf("Failed to create namespace: %v\n", err)
	}

	AwsSecretName = createAwsCredentials()
})

var _ = AfterSuite(func() {
	if err := Operator.Uninstall(); err != nil {
		fmt.Printf("Failed to uninstall operator: %v\n", err)
	}
	if err := kubernetes.DeleteNamespace(Client, TestNamespace); err != nil {
		fmt.Printf("Failed to delete namespace: %v\n", err)
	}

	if Secret != nil {
		if err := Secret.Delete(); err != nil {
			fmt.Printf("Error while deleting AWS secret")
		}
	}
	if TlsSecret != nil {
		if err := TlsSecret.Delete(); err != nil {
			fmt.Printf("Error while deleting TLS secret")
		}
	}
	if AuthSecret != nil {
		if err := AuthSecret.Delete(); err != nil {
			fmt.Printf("Error while deleting Auth secret")
		}
	}

	if err := aws.DeleteFolderInS3(BackupBucket, BackupPrefix); err != nil {
		fmt.Printf("Error while cleaning up S3 bucket for non-tls: %v\n", err)
	}
	if err := aws.DeleteFolderInS3(BackupBucket, BackupPrefixTls); err != nil {
		fmt.Printf("Error while cleaning up S3 bucket for tls: %v\n", err)
	}
})

var _ = AfterEach(func() {
	_ = debug.CollectArtifacts(Client, afero.NewOsFs(), GinkgoWriter, TestNamespace, KubectlPath)
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}

func createTlsSecret() string {
	const secretName = "cassandra-tls"
	s, err := tls.CreateCertSecret(secretName).
		WithNamespace(TestNamespace).
		WithCommonName("CassandraCA").
		Do(Client)

	Expect(err).NotTo(HaveOccurred())

	TlsSecret = &s
	return secretName
}

func createAuthSecret() string {
	const secretName = "authn-credentials" ////nolint:gosec
	s, err := kubernetes.CreateSecret(secretName).
		WithNamespace(TestNamespace).
		WithStringData(map[string]string{
			"username": "cassandra",
			"password": "cassandra",
		}).
		Do(Client)

	Expect(err).NotTo(HaveOccurred())

	AuthSecret = &s
	return secretName
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
	awsSecret, err := kubernetes.CreateSecret(awsSecretName).
		WithNamespace(TestNamespace).
		WithStringData(awsCredentials).Do(Client)

	Expect(err).NotTo(HaveOccurred())

	Secret = &awsSecret
	return awsSecretName
}

func printBackupPodLogs() {
	pods, err := kubernetes.ListPods(Client, TestNamespace)
	if err != nil {
		fmt.Printf("Failed to list pods: %v\n", err)
	} else {
		for _, p := range pods {
			log, err2 := p.ContainerLogs("backup")
			if err2 != nil {
				fmt.Printf("Failed to get log for container 'backup' from pod %s\n", p.Name)
			} else {
				fmt.Printf("Log for container 'backup' from pod %s:\n%s\n", p.Name, log)
			}
		}
	}
}

var _ = Describe("backup and restore", func() {

	It("Creates and restores a backup with local JMX and no SSL", func() {
		var err error

		parameters := map[string]string{
			"NODE_COUNT":                             strconv.Itoa(NodeCount),
			"JMX_LOCAL_ONLY":                         "true",
			"BACKUP_RESTORE_ENABLED":                 "true",
			"BACKUP_AWS_CREDENTIALS_SECRET":          AwsSecretName,
			"BACKUP_PREFIX":                          BackupPrefix,
			"BACKUP_NAME":                            BackupName,
			"BACKUP_AWS_S3_BUCKET_NAME":              BackupBucket,
			"POD_MANAGEMENT_POLICY":                  "OrderedReady",
			"NODE_READINESS_PROBE_INITIAL_DELAY_S":   "30",
			"NODE_READINESS_PROBE_PERIOD_S":          "10",
			"NODE_READINESS_PROBE_FAILURE_THRESHOLD": "6",
			"NODE_LIVENESS_PROBE_INITIAL_DELAY_S":    "60",
			"NODE_LIVENESS_PROBE_FAILURE_THRESHOLD":  "6",
		}
		suites.SetSuitesParameters(parameters)

		By("Installing the operator from current directory")
		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(parameters).
			Do(Client)
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		By("Writing Data to the cassandra cluster")
		output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput))

		By("Running the backup plan")
		err = Operator.Instance.UpdateParameters(map[string]string{
			"BACKUP_TRIGGER": "2",
		})
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("backup")
		printBackupPodLogs()
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
		parameters = map[string]string{
			"NODE_COUNT":                             strconv.Itoa(NodeCount),
			"JMX_LOCAL_ONLY":                         "true",
			"BACKUP_RESTORE_ENABLED":                 "true",
			"BACKUP_AWS_CREDENTIALS_SECRET":          AwsSecretName,
			"BACKUP_PREFIX":                          BackupPrefix,
			"BACKUP_NAME":                            BackupName,
			"BACKUP_AWS_S3_BUCKET_NAME":              BackupBucket,
			"RESTORE_FLAG":                           "true",
			"RESTORE_OLD_NAMESPACE":                  TestNamespace,
			"RESTORE_OLD_NAME":                       TestInstance,
			"POD_MANAGEMENT_POLICY":                  "Parallel",
			"NODE_READINESS_PROBE_INITIAL_DELAY_S":   "30",
			"NODE_READINESS_PROBE_PERIOD_S":          "10",
			"NODE_READINESS_PROBE_FAILURE_THRESHOLD": "6",
			"NODE_LIVENESS_PROBE_INITIAL_DELAY_S":    "60",
			"NODE_LIVENESS_PROBE_FAILURE_THRESHOLD":  "6",
		}
		suites.SetSuitesParameters(parameters)

		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(RestoreInstance).
			WithParameters(parameters).
			Do(Client)

		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		By("Reading Data from the cassandra cluster")
		output, err = cassandra.Cqlsh(Client, Operator.Instance, confirmCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput))

		By("Writing Additional Data to the cassandra cluster")
		output, err = cassandra.Cqlsh(Client, Operator.Instance, additionalDataCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput2))

		By("Updating a parameter and trigger a pod restart")
		parameters = map[string]string{
			"NODE_READINESS_PROBE_INITIAL_DELAY_S": "10",
		}
		suites.SetSuitesParameters(parameters)

		err = Operator.Instance.UpdateParameters(parameters)

		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())
		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		By("Reading Data from the cassandra cluster again")
		output, err = cassandra.Cqlsh(Client, Operator.Instance, confirmCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput2))
	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
		Eventually(func() int {
			pods, _ := kubernetes.ListPods(Client, TestNamespace)
			fmt.Printf("Polling pods: %v\n", len(pods))
			return len(pods)
		}, "300s", "10s").Should(Equal(0))
	})

	It("Creates and restores a backup with JMX SSL and authentication", func() {
		var err error

		tlsSecretName := createTlsSecret()
		authSecretName := createAuthSecret()

		parameters := map[string]string{
			"NODE_COUNT":                             strconv.Itoa(NodeCount),
			"AUTHENTICATOR":                          "PasswordAuthenticator",
			"AUTHENTICATION_SECRET_NAME":             authSecretName,
			"JMX_LOCAL_ONLY":                         "false",
			"TLS_SECRET_NAME":                        tlsSecretName,
			"BACKUP_RESTORE_ENABLED":                 "true",
			"BACKUP_AWS_CREDENTIALS_SECRET":          AwsSecretName,
			"BACKUP_PREFIX":                          BackupPrefixTls,
			"BACKUP_NAME":                            BackupName,
			"BACKUP_AWS_S3_BUCKET_NAME":              BackupBucket,
			"POD_MANAGEMENT_POLICY":                  "Parallel",
			"NODE_READINESS_PROBE_INITIAL_DELAY_S":   "30",
			"NODE_READINESS_PROBE_PERIOD_S":          "10",
			"NODE_READINESS_PROBE_FAILURE_THRESHOLD": "6",
			"NODE_LIVENESS_PROBE_INITIAL_DELAY_S":    "60",
			"NODE_LIVENESS_PROBE_FAILURE_THRESHOLD":  "6",
		}
		suites.SetSuitesParameters(parameters)

		By("Installing the operator from current directory")
		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(TestInstance).
			WithParameters(parameters).
			Do(Client)
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		By("Writing Data to the cassandra cluster")
		output, err := cassandra.Cqlsh(Client, Operator.Instance, testCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput))

		By("Running the backup plan")
		err = Operator.Instance.UpdateParameters(map[string]string{
			"BACKUP_TRIGGER": "2",
		})
		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("backup")
		printBackupPodLogs()
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
		parameters = map[string]string{
			"NODE_COUNT":                             strconv.Itoa(NodeCount),
			"AUTHENTICATOR":                          "PasswordAuthenticator",
			"AUTHENTICATION_SECRET_NAME":             authSecretName,
			"JMX_LOCAL_ONLY":                         "false",
			"TLS_SECRET_NAME":                        tlsSecretName,
			"BACKUP_RESTORE_ENABLED":                 "true",
			"BACKUP_AWS_CREDENTIALS_SECRET":          AwsSecretName,
			"BACKUP_PREFIX":                          BackupPrefixTls,
			"BACKUP_NAME":                            BackupName,
			"BACKUP_AWS_S3_BUCKET_NAME":              BackupBucket,
			"RESTORE_FLAG":                           "true",
			"RESTORE_OLD_NAMESPACE":                  TestNamespace,
			"RESTORE_OLD_NAME":                       TestInstance,
			"POD_MANAGEMENT_POLICY":                  "Parallel",
			"NODE_READINESS_PROBE_INITIAL_DELAY_S":   "30",
			"NODE_READINESS_PROBE_PERIOD_S":          "10",
			"NODE_READINESS_PROBE_FAILURE_THRESHOLD": "6",
			"NODE_LIVENESS_PROBE_INITIAL_DELAY_S":    "60",
			"NODE_LIVENESS_PROBE_FAILURE_THRESHOLD":  "6",
		}
		suites.SetSuitesParameters(parameters)

		Operator, err = kudo.InstallOperator(OperatorDirectory).
			WithNamespace(TestNamespace).
			WithInstance(RestoreInstance).
			WithParameters(parameters).
			Do(Client)

		Expect(err).To(BeNil())

		By("Waiting for the plan to complete")
		err = Operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
		Expect(err).To(BeNil())

		suites.AssertNumberOfCassandraNodes(Client, Operator, NodeCount)

		By("Reading Data from the cassandra cluster")
		output, err = cassandra.Cqlsh(Client, Operator.Instance, confirmCQLScript)
		Expect(err).To(BeNil())
		Expect(output).To(ContainSubstring(testCQLScriptOutput))
	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
		Eventually(func() int {
			pods, _ := kubernetes.ListPods(Client, TestNamespace)
			fmt.Printf("Polling pods: %v\n", len(pods))
			return len(pods)
		}, "300s", "10s").Should(Equal(0))
	})
})
