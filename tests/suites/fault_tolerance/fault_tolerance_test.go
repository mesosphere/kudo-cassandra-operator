package fault_tolerance

import (
	"fmt"
	"os"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/thoas/go-funk"

	testclient "github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
	"github.com/kudobuilder/test-tools/pkg/kudo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/mesosphere/kudo-cassandra-operator/tests/cassandra"
)

var (
	kubeConfigPath    = os.Getenv("KUBECONFIG")
	operatorName      = os.Getenv("OPERATOR_NAME")
	operatorDirectory = os.Getenv("OPERATOR_DIRECTORY")

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

func getTopology1DatacenterEach2Rack() cassandra.NodeTopology {
	return cassandra.NodeTopology{
		{
			Datacenter:       "dc1",
			DatacenterLabels: map[string]string{
				// For AWS, currently no DC labels
			},
			Nodes:        4,
			RackLabelKey: nodeSelectorDatacenter,
			Racks: []cassandra.TopologyRackItem{
				{
					Rack:           "rac1",
					RackLabelValue: "us-west-2a",
				},
				{
					Rack:           "rac2",
					RackLabelValue: "us-west-2b",
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

func setupRBAC(client testclient.Client) {
	serviceAccount := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeResolverServiceAccount,
			Namespace: testNamespace,
		},
	}
	_, err := kubernetes.NewServiceAccount(client, serviceAccount)
	if !errors.IsAlreadyExists(err) {
		Expect(err).NotTo(HaveOccurred())
	}

	role := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeResolverRole,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list"},
			},
		},
		AggregationRule: nil,
	}
	_, err = kubernetes.NewClusterRole(client, role)
	if !errors.IsAlreadyExists(err) {
		Expect(err).NotTo(HaveOccurred())
	}

	roleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeResolverRoleBinding,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     nodeResolverRole,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      nodeResolverServiceAccount,
			Namespace: testNamespace,
		}},
	}
	_, err = kubernetes.NewClusterRoleBinding(client, roleBinding)
	if !errors.IsAlreadyExists(err) {
		Expect(err).NotTo(HaveOccurred())
	}

}

var _ = Describe("Fault tolerance tests", func() {

	var (
		client     testclient.Client
		operator   kudo.Operator
		parameters map[string]string
	)

	AfterEach(func() {
		err := operator.Uninstall()
		Expect(err).NotTo(HaveOccurred())

		deleteRBAC(client)

		err = kubernetes.DeleteNamespace(client, testNamespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when configured with the 'GossipingPropertyFileSnitch' snitch", func() {
		It("should set up the datacenter and rack properties", func() {
			var err error

			client, err = testclient.NewForConfig(kubeConfigPath)
			Expect(err).NotTo(HaveOccurred())

			By("Setting up Namespace and RBAC")
			err = kubernetes.CreateNamespace(client, testNamespace)
			if !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
			deleteRBAC(client)
			setupRBAC(client)

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
