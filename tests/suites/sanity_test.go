package suites

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	// log "github.com/sirupsen/logrus"

	cassandra "github.com/mesosphere/kudo-cassandra-operator/tests/utils/cassandra"
	k8s "github.com/mesosphere/kudo-cassandra-operator/tests/utils/k8s"
	kubectl "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kubectl"
	kudo "github.com/mesosphere/kudo-cassandra-operator/tests/utils/kudo"
)

var (
	TestName          = "sanity-test"
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
	It("Installs the operator from a directory", func() {
		// TODO(mpereira) Assert that it isn't running.
		err := kudo.InstallOperatorFromDirectory(
			OperatorDirectory, TestNamespace, TestInstance, []string{},
		)
		// TODO(mpereira) Assert that it is running.
		if err != nil {
			Fail(
				"Failing the full suite: failed to install operator instance that the " +
					"following tests depend on",
			)
		}
		Expect(err).To(BeNil())
	})

	It("Scales the instance's number of nodes", func() {
		err := kudo.UpdateInstanceParameters(
			TestNamespace, TestInstance, map[string]string{"NODE_COUNT": "4"},
		)
		if err != nil {
			Fail("Failing the full suite: failed to scale the number of nodes")
		}
		Expect(err).To(BeNil())
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
	})

	It("Customize cassandra.yaml", func() {
		parameter := "otc_backlog_expiration_interval_ms"
		initialValue := "200"
		desiredValue := "300"
		desiredValueBase64 := base64.StdEncoding.EncodeToString([]byte("otc_backlog_expiration_interval_ms: 300"))

		configuration, err := cassandra.ClusterConfiguration(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = kudo.UpdateInstanceParameters(
			TestNamespace,
			TestInstance,
			map[string]string{"CUSTOM_CASSANDRA_YAML_BASE64": desiredValueBase64},
		)
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(
			TestNamespace, TestInstance,
		)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))
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
	k8s.DeleteNamespace(TestNamespace)
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf(
		"%s-junit.xml", TestName,
	))
	RunSpecsWithDefaultAndCustomReporters(t, TestName, []Reporter{junitReporter})
}
