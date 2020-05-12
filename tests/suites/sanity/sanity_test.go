package sanity

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/debug"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
	"github.com/mesosphere/kudo-cassandra-operator/tests/curl"
	"github.com/mesosphere/kudo-cassandra-operator/tests/prometheus"
	"github.com/mesosphere/kudo-cassandra-operator/tests/suites"
)

var (
	TestName          = "sanity-test"
	OperatorName      = os.Getenv("OPERATOR_NAME")
	TestNamespace     = fmt.Sprintf("%s-namespace", TestName)
	TestInstance      = fmt.Sprintf("%s-instance", OperatorName)
	KubeConfigPath    = os.Getenv("KUBECONFIG")
	KubectlPath       = os.Getenv("KUBECTL_PATH")
	OperatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

	// Node Count of 1 for Sanity test to have the tests a little bit faster
	NodeCount = 1
	Client    = client.Client{}
	Operator  = kudo.Operator{}
	Fs        = afero.NewOsFs()
)

var _ = BeforeSuite(func() {
	Client, _ = client.NewForConfig(KubeConfigPath)
	_ = kubernetes.CreateNamespace(Client, TestNamespace)
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

func assertNumberOfCassandraNodes(nodeCount int) {
	nodes, err := cassandra.Nodes(Client, Operator.Instance)
	Expect(err).To(BeNil())
	Expect(len(nodes)).To(Equal(nodeCount))
}

var _ = Describe(TestName, func() {
	It("Installs the current operator", func() {
		var err error

		parameters := map[string]string{
			"NODE_COUNT": strconv.Itoa(NodeCount),
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
		assertNumberOfCassandraNodes(NodeCount)

		if !suites.IsLocalCluster() {
			By("providing metrics to prometheus")
			prometheusSvc := "prometheus-kubeaddons-prom-prometheus.kubeaddons.svc.cluster.local:9090"

			curlRunner := curl.New(Client, TestNamespace)

			Eventually(func() bool {
				promResult, err := prometheus.QueryForStats(curlRunner, prometheusSvc, "cassandra_stats")
				Expect(err).To(BeNil())

				return len(promResult.Data.Result) > 0
			}, 5*time.Minute, 30*time.Second).Should(BeTrue())
		}

		By("Updating the instances cpu and memory")
		newMemMiB := 3192
		newMemLimitMiB := 3192
		newMemBytes := newMemMiB * 1024 * 1024
		newMemLimitBytes := newMemLimitMiB * 1024 * 1024

		newCPU := 800
		newCPULimit := 1100

		err = Operator.Instance.UpdateParameters(map[string]string{
			"NODE_MEM_MIB":       strconv.Itoa(newMemMiB),
			"NODE_MEM_LIMIT_MIB": strconv.Itoa(newMemLimitMiB),
			"NODE_CPU_MC":        strconv.Itoa(newCPU),
			"NODE_CPU_LIMIT_MC":  strconv.Itoa(newCPULimit),
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		pod, err := kubernetes.GetPod(Client, TestInstance+"-node-0", TestNamespace)

		Expect(err).To(BeNil())
		Expect(pod).To(Not(BeNil()))

		Expect(pod.Spec.Containers[0].Resources.Requests.Cpu().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newCPU))))
		Expect(pod.Spec.Containers[0].Resources.Requests.Memory().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newMemBytes))))

		Expect(pod.Spec.Containers[0].Resources.Limits.Cpu().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newCPULimit))))
		Expect(pod.Spec.Containers[0].Resources.Limits.Memory().AsDec().UnscaledBig()).To(Equal(big.NewInt(int64(newMemLimitBytes))))

		assertNumberOfCassandraNodes(NodeCount)

		By("Updating the instances parameters")
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

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)

		By("Configuring Cassandras properties through custom properties")
		parameter = "otc_backlog_expiration_interval_ms"
		initialValue = "200"
		desiredValue = "300"
		desiredEncodedProperties := base64.StdEncoding.EncodeToString(
			[]byte(parameter + ": " + desiredValue),
		)

		configuration, err = cassandra.ClusterConfiguration(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = Operator.Instance.UpdateParameters(map[string]string{
			"CUSTOM_CASSANDRA_YAML_BASE64": desiredEncodedProperties,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		configuration, err = cassandra.ClusterConfiguration(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)

		By("Configuring Cassandra JVM options through custom options")
		parameter = "-XX:CMSWaitDuration"
		initialValue = "10000"
		desiredValue = "11000"
		desiredEncodedProperties = base64.StdEncoding.EncodeToString(
			[]byte(parameter + "=" + desiredValue),
		)

		configuration, err = cassandra.NodeJVMOptions(Client, Operator.Instance)

		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(initialValue))

		err = Operator.Instance.UpdateParameters(map[string]string{
			"CUSTOM_JVM_OPTIONS_BASE64": desiredEncodedProperties,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		configuration, err = cassandra.NodeJVMOptions(Client, Operator.Instance)
		Expect(err).To(BeNil())
		Expect(configuration[parameter]).To(Equal(desiredValue))

		assertNumberOfCassandraNodes(NodeCount)

		By("Scaling the instances number of nodes")
		// Make sure we create an actual cluster of three nodes
		NodeCount = NodeCount + 2
		err = Operator.Instance.UpdateParameters(map[string]string{
			"NODE_COUNT": strconv.Itoa(NodeCount)},
		)
		if err != nil {
			Fail("Failing the full suite: failed to scale the number of nodes")
		}
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("deploy")
		Expect(err).To(BeNil())

		assertNumberOfCassandraNodes(NodeCount)

		By("Triggering a Cassandra node repair")
		podName, err := cassandra.PodName(Operator.Instance, 0)
		Expect(err).To(BeNil())

		err = Operator.Instance.UpdateParameters(map[string]string{
			"REPAIR_POD": podName,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("repair")
		Expect(err).To(BeNil())

		repair, err := cassandra.NodeWasRepaired(Client, Operator.Instance, podName)
		Expect(err).To(BeNil())
		Expect(repair).To(BeTrue())

		By("Triggering a Cassandra node repair again to make sure it can execute twice")
		podName, err = cassandra.PodName(Operator.Instance, 1)
		Expect(err).To(BeNil())

		err = Operator.Instance.UpdateParameters(map[string]string{
			"REPAIR_POD": podName,
		})
		Expect(err).To(BeNil())

		err = Operator.Instance.WaitForPlanComplete("repair")
		Expect(err).To(BeNil())

		repair, err = cassandra.NodeWasRepaired(Client, Operator.Instance, podName)
		Expect(err).To(BeNil())
		Expect(repair).To(BeTrue())

		By("Triggering a Cassandra node repair on an invalid pod name")
		podName = "invalid-pod"
		Expect(err).To(BeNil())

		err = Operator.Instance.UpdateParameters(map[string]string{
			"REPAIR_POD": podName,
		})
		Expect(err).To(BeNil())

		err = WaitForPlanFailure(Operator.Instance, "repair")
		Expect(err).To(BeNil())
	})

	It("Uninstalls the operator", func() {
		err := cassandra.Uninstall(Client, Operator)
		Expect(err).To(BeNil())
		// TODO(mpereira) Assert that it isn't running.
	})
})

func WaitForPlanFailure(instance kudo.Instance, plan string, options ...kudo.WaitOption) error {
	config := kudo.WaitConfig{
		Timeout: time.Minute * 5,
	}

	for _, option := range options {
		option(&config)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), config.Timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	return instance.WaitForPlanStatus(ctx, ticker, plan, v1beta1.ExecutionFatalError)
}
