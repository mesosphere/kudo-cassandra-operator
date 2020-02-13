package sanity

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/prometheus"
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
		prometheusSvc := "prometheus-kubeaddons-prom-prometheus.kubeaddons.svc.cluster.local:9090"

		Eventually(func() bool {
			promResult, err := prometheus.QueryForStats(Client, TestNamespace, prometheusSvc, "cassandra_stats")
			Expect(err).To(BeNil())

			return len(promResult.Data.Result) > 0
		}, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})

	It("Updates the instance's cpu and memory", func() {
		newMemMiB := 3192
		newMemLimitMiB := 3192
		newMemBytes := newMemMiB * 1024 * 1024
		newMemLimitBytes := newMemLimitMiB * 1024 * 1024

		newCpu := 800
		newCpuLimit := 1100

		err := Operator.Instance.UpdateParameters(map[string]string{
			"NODE_MEM_MIB":       strconv.Itoa(newMemMiB),
			"NODE_MEM_LIMIT_MIB": strconv.Itoa(newMemLimitMiB),
			"NODE_CPU_MC":        strconv.Itoa(newCpu),
			"NODE_CPU_LIMIT_MC":  strconv.Itoa(newCpuLimit),
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanInProgress("deploy")
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		pod, err := kubernetes.GetPod(Client, TestInstance+"-node-0", TestNamespace)

		Expect(err).To(BeNil())
		Expect(pod).To(Not(BeNil()))

		Expect(pod.Spec.Containers[0].Resources.Requests.Cpu().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newCpu))))
		Expect(pod.Spec.Containers[0].Resources.Requests.Memory().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newMemBytes))))

		Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newCpuLimit))))
		Expect(pod.Spec.Containers[0].Resources.Limits.Memory().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newMemLimitBytes))))

		assertNumberOfCassandraNodes(NodeCount)
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

		// Make sure we create an actual cluster of three nodes
		NodeCount = NodeCount + 2
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
	_ = kubernetes.CreateNamespace(Client, TestNamespace)
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
