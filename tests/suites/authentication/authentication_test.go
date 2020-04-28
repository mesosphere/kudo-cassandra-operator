package authentication

import (
	"fmt"
	"os"
	"testing"
	"time"

	testclient "github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	kubeConfigPath    = os.Getenv("KUBECONFIG")
	operatorName      = os.Getenv("OPERATOR_NAME")
	operatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	instanceName  = fmt.Sprintf("%s-instance", operatorName)
	testNamespace = "authentication"
)

var _ = Describe("Authentication tests", func() {
	var (
		client      testclient.Client
		credentials kubernetes.Secret
		operator    kudo.Operator
	)

	AfterEach(func() {
		err := operator.Uninstall()
		Expect(err).NotTo(HaveOccurred())

		err = credentials.Delete()
		Expect(err).NotTo(HaveOccurred())

		err = kubernetes.DeleteNamespace(client, testNamespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when using the 'PasswordAuthenticator'", func() {
		It("should authenticate 'nodetool' calls", func() {
			var err error

			client, err = testclient.NewForConfig(kubeConfigPath)
			Expect(err).NotTo(HaveOccurred())

			By("Setting up namespace")
			err = kubernetes.CreateNamespace(client, testNamespace)
			if !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			By("Adding a secret containing the default user credentials")
			const secretName = "authn-credentials" ////nolint:gosec

			credentials, err = kubernetes.CreateSecret(secretName).
				WithNamespace(testNamespace).
				WithStringData(map[string]string{
					"username": "cassandra",
					"password": "cassandra",
				}).
				Do(client)
			Expect(err).NotTo(HaveOccurred())

			By("Installing the operator with 'PasswordAuthenticator'")
			parameters := map[string]string{
				"AUTHENTICATOR":              "PasswordAuthenticator",
				"AUTHENTICATION_SECRET_NAME": secretName,
			}

			operator, err = kudo.InstallOperator(operatorDirectory).
				WithNamespace(testNamespace).
				WithInstance(instanceName).
				WithParameters(parameters).
				Do(client)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			By("Triggering a Cassandra node repair which uses 'nodetool'")
			podName, err := cassandra.FirstPodName(operator.Instance)
			Expect(err).To(BeNil())

			err = operator.Instance.UpdateParameters(map[string]string{
				"REPAIR_POD": podName,
			})
			Expect(err).To(BeNil())

			err = operator.Instance.WaitForPlanComplete("repair-pod")
			Expect(err).To(BeNil())

			repair, err := cassandra.NodeWasRepaired(client, operator.Instance, podName)
			Expect(err).To(BeNil())
			Expect(repair).To(BeTrue())
		})
	})
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("authentication-test-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Authentication tests", []Reporter{junitReporter})
}
