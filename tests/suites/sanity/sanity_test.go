package sanity

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/cmd"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
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
				"NODE_COUNT": strconv.Itoa(NodeCount),
			}).
			Do(Client)
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Upgrades the running operator instance from a directory", func() {
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

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("provides metrics to prometheus", func() {
		podTpl := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "curl-pod",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "curl",
						Image:   "curlimages/curl",
						Command: []string{"sleep", "300"},
					},
				},
			},
		}

		pod, err := kubernetes.NewPod(Client, podTpl)
		Expect(err).To(BeNil())

		for {
			if pod.Status.Phase == v1.PodRunning {
				break
			} else {
				pod, err = kubernetes.GetPod(Client, pod.Name, pod.Namespace)
				Expect(err).To(BeNil())
			}
		}

		var stdout strings.Builder
		var stderr strings.Builder

		cmd := cmd.New("curl").
			WithArguments("-s", "prometheus-kubeaddons-prom-prometheus.kubeaddons.svc.cluster.local:9090/api/v1/query?query=cassandra").
			WithStdout(&stdout).
			WithStderr(&stderr)

		err = pod.ContainerExec("curl", cmd)
		Expect(err).To(BeNil())

		log.Info("Curl Output:\n StdErr: %v\n StdOut: %v\n", stderr.String(), stdout.String())

		expectedOutput := `{"status":"success","data":{"resultType":"vector","result":[]}}`
		Expect(stdout.String()).To(Equal(expectedOutput))

		err = pod.Delete()
		Expect(err).To(BeNil())
	})

	It("Updates the instance's parameters", func() {
		parameter := "disk_failure_policy"
		initialValue := "stop"
		desiredValue := "ignore"

		configuration, err := cassandra.ClusterConfiguration(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = Operator.Instance.UpdateParameters(map[string]string{
			strings.ToUpper(parameter): desiredValue,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(Client, Operator.Instance)
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

		configuration, err := cassandra.ClusterConfiguration(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = Operator.Instance.UpdateParameters(map[string]string{
			"CUSTOM_CASSANDRA_YAML_BASE64": desiredEncodedProperties,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(Client, Operator.Instance)
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

		configuration, err := cassandra.NodeJVMOptions(Client, Operator.Instance)

		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = Operator.Instance.UpdateParameters(map[string]string{
			"CUSTOM_JVM_OPTIONS_BASE64": desiredEncodedProperties,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		configuration, err = cassandra.NodeJVMOptions(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Scales the instance's number of nodes", func() {
		NodeCount = NodeCount + 1
		err := Operator.Instance.UpdateParameters(map[string]string{
			"NODE_COUNT": strconv.Itoa(NodeCount)},
		)
		if err != nil {
			Fail("Failing the full suite: failed to scale the number of nodes")
		}
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)
	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
		// TODO(mpereira) Assert that it isn't running.
	})
})

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	kubernetes.CreateNamespace(Client, TestNamespace)
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
