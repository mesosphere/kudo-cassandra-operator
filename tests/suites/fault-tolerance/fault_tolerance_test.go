package faulttolerance

import (
	"fmt"
	"os"
	"testing"
	"time"

	testclient "github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/debug"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	kubeConfigPath    = os.Getenv("KUBECONFIG")
	kubectlPath       = os.Getenv("KUBECTL_PATH")
	operatorName      = os.Getenv("OPERATOR_NAME")
	operatorDirectory = os.Getenv("OPERATOR_DIRECTORY")
	client            testclient.Client
	operator          kudo.Operator

	instanceName  = fmt.Sprintf("%s-instance", operatorName)
	testNamespace = "fault-tolerance"

	// This label on the nodes is used to distinguish datacenters
	nodeSelectorDatacenter = "failure-domain.beta.kubernetes.io/zone"

	// This label on the nodes is used to distinguish racks (Not used on AWS)
	//nodeSelectorRack  	   = "failure-domain.beta.kubernetes.io/region"

	// We need some node label/value pair to select racks. As we don't have any specific
	// rack labels on the TC cluster, we use the region. This is not perfect, but
	// should work for now.
	rackLabelKey   = "failure-domain.beta.kubernetes.io/region"
	rackLabelValue = "us-west-2"

	// RBAC names for Role, RoleBinding and service account
	nodeResolverServiceAccount = "node-resolver"
	nodeResolverRole           = "node-resolver-role"
	nodeResolverRoleBinding    = "node-resolver-rolebinding"
)

const createSchemaCQLTemplate = "CREATE SCHEMA schema1 WITH replication = { 'class' : 'NetworkTopologyStrategy', %s };"

const testCQLScript = "USE schema1;" +
	"CREATE TABLE users (user_id varchar PRIMARY KEY,first varchar,last varchar,age int);" +
	"INSERT INTO users (user_id, first, last, age) VALUES ('jsmith', 'John', 'Smith', 42);" +
	"SELECT * FROM users;"

const testCQLOutputScript = "USE schema1;" +
	"SELECT * FROM users;"

const testCQLScriptOutput = `
 user_id | age | first | last
---------+-----+-------+-------
  jsmith |  42 |  John | Smith

(1 rows)`

func buildDatacenterReplicationString(topology cassandra.NodeTopology, maxReplica int) string {
	result := ""
	for _, datacenter := range topology {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("'%s': %d", datacenter.Datacenter, funk.MinInt([]int{maxReplica, datacenter.Nodes}))
	}
	return result
}

func getTopology1DatacenterEach1Rack(datacenter, rack string) cassandra.NodeTopology {
	return cassandra.NodeTopology{
		{
			Datacenter: datacenter,
			DatacenterLabels: map[string]string{
				nodeSelectorDatacenter: "us-west-2a",
			},
			Nodes:        1,
			RackLabelKey: rackLabelKey,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           rack,
					RackLabelValue: rackLabelValue,
				},
			},
		},
	}
}

func getTopology2DatacenterEach1Rack() cassandra.NodeTopology {
	return cassandra.NodeTopology{
		{
			Datacenter: "dc1",
			DatacenterLabels: map[string]string{
				nodeSelectorDatacenter: "us-west-2a",
			},
			Nodes:        2,
			RackLabelKey: rackLabelKey,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           "rac1",
					RackLabelValue: rackLabelValue,
				},
			},
		},
		{
			Datacenter: "dc2",
			DatacenterLabels: map[string]string{
				nodeSelectorDatacenter: "us-west-2b",
			},
			Nodes:        2,
			RackLabelKey: rackLabelKey,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           "rac1",
					RackLabelValue: rackLabelValue,
				},
			},
		},
	}
}

func getTopology3DatacenterEach1Rack() cassandra.NodeTopology {
	return cassandra.NodeTopology{
		{
			Datacenter: "dc1",
			DatacenterLabels: map[string]string{
				nodeSelectorDatacenter: "us-west-2a",
			},
			Nodes:        2,
			RackLabelKey: rackLabelKey,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           "rac1",
					RackLabelValue: rackLabelValue,
				},
			},
		},
		{
			Datacenter: "dc2",
			DatacenterLabels: map[string]string{
				nodeSelectorDatacenter: "us-west-2b",
			},
			Nodes:        2,
			RackLabelKey: rackLabelKey,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           "rac1",
					RackLabelValue: rackLabelValue,
				},
			},
		},
		{
			Datacenter: "dc3",
			DatacenterLabels: map[string]string{
				nodeSelectorDatacenter: "us-west-2c",
			},
			Nodes:        2,
			RackLabelKey: rackLabelKey,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           "rac1",
					RackLabelValue: rackLabelValue,
				},
			},
		},
	}
}

func deleteRBAC(client testclient.Client) {
	roleBinding, err := kubernetes.GetClusterRoleBinding(client, nodeResolverRoleBinding)
	if err == nil {
		err := roleBinding.Delete()
		if err != nil {
			fmt.Printf("Failed to delete ClusterRoleBinding: %v", err)
		}
	}

	role, err := kubernetes.GetClusterRole(client, nodeResolverRole)
	if err == nil {
		err := role.Delete()
		if err != nil {
			fmt.Printf("Failed to delete ClusterRole: %v", err)
		}
	}

	serviceAccount, err := kubernetes.GetServiceAccount(client, nodeResolverServiceAccount, testNamespace)
	if err == nil {
		err := serviceAccount.Delete()
		if err != nil {
			fmt.Printf("Failed to delete ServiceAccount: %v", err)
		}
	}
}

var _ = BeforeEach(func() {
	var err error

	client, err = testclient.NewForConfig(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	debug.CollectArtifacts(client, afero.NewOsFs(), GinkgoWriter, testNamespace, kubectlPath)

	err := operator.Uninstall()
	Expect(err).NotTo(HaveOccurred())

	deleteRBAC(client)
})

var _ = AfterSuite(func() {
	err := kubernetes.DeleteNamespace(client, testNamespace)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Fault tolerance tests", func() {

	var parameters map[string]string

	Context("when configured with the 'GossipingPropertyFileSnitch' snitch", func() {
		It("should set up the datacenter and rack properties", func() {
			var err error

			By("Setting up Namespace and RBAC")
			err = kubernetes.CreateNamespace(client, testNamespace)
			if !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
			deleteRBAC(client)

			By("Starting the test")

			By("Installing the operator with a topology")
			topology := getTopology2DatacenterEach1Rack()
			topologyYaml, err := topology.ToYAML()
			Expect(err).NotTo(HaveOccurred())

			parameters = map[string]string{
				"NODE_COUNT":                           "1", // NODE_TOPOLOGY should override this value
				"ENDPOINT_SNITCH":                      "GossipingPropertyFileSnitch",
				"NODE_TOPOLOGY":                        topologyYaml,
				"NODE_ANTI_AFFINITY":                   "true",
				"NODE_READINESS_PROBE_INITIAL_DELAY_S": "10",
				"SERVICE_ACCOUNT_INSTALL":              "true",
			}

			By("Waiting for the operator to deploy")
			operator, err = kudo.InstallOperator(operatorDirectory).
				WithNamespace(testNamespace).
				WithInstance(instanceName).
				WithParameters(parameters).
				Do(client)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			By("Ensuring that all nodes are up")

			nodes, err := cassandra.Nodes(client, operator.Instance)
			Expect(err).NotTo(HaveOccurred())

			var totalNodes = 0
			for _, dc := range topology {
				totalNodes += dc.Nodes
			}
			Expect(nodes).To(HaveLen(totalNodes))

			By("Writing data to the cluster")
			replicationString := buildDatacenterReplicationString(topology, 2)
			createSchemaCQL := fmt.Sprintf(createSchemaCQLTemplate, replicationString)

			output, err := cassandra.Cqlsh(client, operator.Instance, createSchemaCQL+testCQLScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))

			By("Checking that node status reports correct data center")
			dcCounts := collectDataCenterCounts(nodes)
			for _, dc := range topology {
				Expect(dcCounts[dc.Datacenter]).To(Equal(dc.Nodes))
			}

			By("Updating the topology")
			topology = getTopology3DatacenterEach1Rack()
			topologyYaml, err = topology.ToYAML()
			Expect(err).To(BeNil())
			parameters = map[string]string{
				"NODE_TOPOLOGY": topologyYaml,
			}
			err = operator.Instance.UpdateParameters(parameters)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*15))
			Expect(err).NotTo(HaveOccurred())

			By("Ensuring that all nodes are up")

			nodes, err = cassandra.Nodes(client, operator.Instance)
			Expect(err).NotTo(HaveOccurred())

			dcCounts = collectDataCenterCounts(nodes)
			for _, dc := range topology {
				Expect(dcCounts[dc.Datacenter]).To(Equal(dc.Nodes))
			}

			By("Reading data from the cluster")
			output, err = cassandra.Cqlsh(client, operator.Instance, testCQLOutputScript)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(testCQLScriptOutput))

			By("Checking that nodes are deployed on different ndoes")
			podList, err := kubernetes.ListPods(client, testNamespace)
			Expect(err).To(BeNil())

			usedIPs := map[string]bool{}
			for _, pod := range podList {
				ip := pod.Status.HostIP
				_, ok := usedIPs[ip]
				Expect(ok).To(BeFalse(), "HostIP has been reused, anti-affinity should prevent that")
			}
		})

		// TODO: test node selection
	})

	Context("when having two datacenters in different namespaces", func() {
		var (
			dc2Namespace = "fault-tolerance-2"
		)

		It("is recognized by the respective clusters", func() {
			var err error

			By("Setting up Namespace and RBAC")
			err = kubernetes.CreateNamespace(client, testNamespace)
			if !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			err = kubernetes.CreateNamespace(client, dc2Namespace)
			if !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			defer func() {
				err := kubernetes.DeleteNamespace(client, dc2Namespace)
				Expect(err).NotTo(HaveOccurred())
			}()

			By("Starting the test")

			By("Installing an operator in the first namespace")
			topology := getTopology1DatacenterEach1Rack("dc1", "rac1")
			topologyYaml, err := topology.ToYAML()
			Expect(err).NotTo(HaveOccurred())

			parameters = map[string]string{
				"NODE_COUNT":                           "1", // NODE_TOPOLOGY should override this value
				"ENDPOINT_SNITCH":                      "GossipingPropertyFileSnitch",
				"NODE_TOPOLOGY":                        topologyYaml,
				"NODE_READINESS_PROBE_INITIAL_DELAY_S": "10",
				"SERVICE_ACCOUNT_INSTALL":              "true",
			}

			By("Waiting for the operator to deploy")
			operator, err = kudo.InstallOperator(operatorDirectory).
				WithNamespace(testNamespace).
				WithInstance(instanceName).
				WithParameters(parameters).
				Do(client)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			By("Installing an operator in the second namespace with external seed node from the first operator")
			topology = getTopology1DatacenterEach1Rack("dc2", "rac2")
			topologyYaml, err = topology.ToYAML()
			Expect(err).NotTo(HaveOccurred())

			dns := fmt.Sprintf("[%s-dc1-node-0.%s-svc.%s.cluster.local]", instanceName, instanceName, testNamespace)

			parameters = map[string]string{
				"NODE_COUNT":                           "1", // NODE_TOPOLOGY should override this value
				"ENDPOINT_SNITCH":                      "GossipingPropertyFileSnitch",
				"NODE_TOPOLOGY":                        topologyYaml,
				"NODE_READINESS_PROBE_INITIAL_DELAY_S": "10",
				"SERVICE_ACCOUNT_INSTALL":              "true",
				"EXTERNAL_SEED_NODES":                  dns,
			}

			By("Waiting for the second operator to deploy")
			operator2, err := kudo.InstallOperator(operatorDirectory).
				WithNamespace(dc2Namespace).
				WithInstance(instanceName).
				WithParameters(parameters).
				Do(client)
			Expect(err).NotTo(HaveOccurred())

			err = operator2.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			nodes, err := cassandra.Nodes(client, operator2.Instance)
			Expect(err).NotTo(HaveOccurred())

			dcCounts := collectDataCenterCounts(nodes)
			Expect(dcCounts["dc1"] == 1)
			Expect(dcCounts["dc2"] == 1)

			By("Updating the external seed nodes of the first operator")
			dns = fmt.Sprintf("[%s-dc2-node-0.%s-svc.%s.cluster.local]", instanceName, instanceName, dc2Namespace)

			parameters = map[string]string{
				"EXTERNAL_SEED_NODES": dns,
			}

			err = operator.Instance.UpdateParameters(parameters)
			Expect(err).NotTo(HaveOccurred())

			err = operator.Instance.WaitForPlanComplete("deploy", kudo.WaitTimeout(time.Minute*10))
			Expect(err).NotTo(HaveOccurred())

			nodes, err = cassandra.Nodes(client, operator.Instance)
			Expect(err).NotTo(HaveOccurred())

			dcCounts = collectDataCenterCounts(nodes)
			Expect(dcCounts["dc1"] == 1)
			Expect(dcCounts["dc2"] == 1)
		})
	})
})

func collectDataCenterCounts(nodes []map[string]string) map[string]int {
	result := map[string]int{}
	for _, node := range nodes {
		val, ok := result[node["datacenter"]]
		if !ok {
			result[node["datacenter"]] = 1
		} else {
			result[node["datacenter"]] = val + 1
		}
	}
	return result
}

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("fault-tolerance-test-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Fault tolerance tests", []Reporter{junitReporter})
}
