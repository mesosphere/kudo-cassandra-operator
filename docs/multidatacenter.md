# KUDO Cassandra with Multiple Datacenters and Rack awareness

This guide explains the details of a multi-datacenter setup for KUDO Cassandra

## Description

Cassandra supports different topologies, including different datacenters and rack awareness.

- Different datacenters usually provide complete replication and locality. Each datacenter usually contains a separate 'ring'
- Different racks indicate different failure zones to Cassandra: Data is replicated in a way that different copies are not stored in the same rack.


## Kubernetes cluster prerequisites

At this time, KUDO Cassandra needs a single Kubernetes cluster spanning all the datacenters. A Cassandra cluster running on two or more Kubernetes
clusters is not supported at the moment

### Node labels

- Datacenter labels: Each Kubernetes node must have appropriate labels indicating the datacenter it belongs to. These labels can have any name, but
it is advised to use the standard [Kubernetes topology labels](https://kubernetes.io/docs/reference/kubernetes-api/labels-annotations-taints/#topologykubernetesioregion).

If the Kubernets cluster is running on AWS, these labels are usually set by default, on AWS they correspond to the different regions:
```
topology.kubernetes.io/region=us-east-1
```

As datacenter selection is configured on datacenter level for cassandra, it is possible to use different keys for each datacenter. This might
be especially useful for hybrid clouds. For example, this would be an valid configuration:
```
Datacenter 1 (OnPrem):
nodeLabels:
  custom.topology=onprem

Datacenter 2 (AWS):
nodeLabels:
  topology.kubernetes.io/region=us-east-1
```

- Rack labels: Additionally to the datacenter label, each kubernetes node must have a rack label. This label defines how Cassandra distributes data in 
each datacenter, and can correspond to AWS availability zones, actual rack names in a datacenter, power groups, etc.

Again, it is advised to use the [Kubernetes topology labels](https://kubernetes.io/docs/reference/kubernetes-api/labels-annotations-taints/#topologykubernetesioregion), for example on
AWS a label would look like:
```
topology.kubernetes.io/zone=us-east-1c
```

The label key is again defined on datacenter level and therefore they key needs to be the same for all nodes used in the same datacenter.

### Service Account and RBAC

As there is currently no easy way to read node labels from inside a pod, the KUDO Cassandra operator uses an initContainer to read the rack of the deployed pod. This requires
a service account with valid RBAC permissions. Please refer to the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) on how to 
create service accounts.

The service account requires the permissions to `get` and `list` the `nodes` resource.

#### Create Cluster Role
```bash
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: node-resolver-role
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "watch", "list"]
EOF
```

#### Create Service Account
```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: node-resolver
  namespace: ${kudo_cassandra_instance_namespace}
EOF
```

#### Create ClusterRoleBinding
```bash
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: node-resolver-binding
subjects:
- kind: ServiceAccount
  name: node-resolver
  namespace: ${kudo_cassandra_instance_namespace}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: node-resolver-role
EOF
```

The service account then needs to be set in the `NODE_RESOLVE_SERVICEACCOUNT` parameter.

## Topology

KUDO Cassandra supports the cluster setup with a single `NODE_TOPOLOGY` parameter. This parameter contains a YAML structure that
describes the expected setup.

An example:
```yaml
- datacenter: dc1
  datacenterLabels:
    failure-domain.beta.kubernetes.io/region: us-west-2
  nodes: 9
  rackLabelKey: failure-domain.beta.kubernetes.io/zone
  racks:
  - rack: rack1
    rackLabelValue: us-west-2a
  - rack: rack2
    rackLabelValue: us-west-2b
  - rack: rack3
    rackLabelValue: us-west-2c
- datacenter: dc2
  datacenterLabels:
    failure-domain.beta.kubernetes.io/region: us-east-1
  nodes: 9
  rackLabelKey: failure-domain.beta.kubernetes.io/zone
  racks:
  - rack: rack4
    rackLabelValue: us-east-1a
  - rack: rack5
    rackLabelValue: us-east-1b
  - rack: rack6
    rackLabelValue: us-east-1c
```

This deployment requires a kubernetes cluster of at least 18 worker nodes, with at least 9 in each `us-west-2` and `us-east1` region.

It will deploy two StatefulSets with each 9 pods. Each StatefulSet creates it's own ring, the replication factor between the datacenters
can be specified on the keyspace level inside cassandra.

It is *not* possible to exactly specify how many pods will be started on each rack at the moment - the KUDO Cassandra operator and Kubernetes
will distribute the Cassandra nodes over all specified racks by availability and with the most possible spread:

For example, if we use the above example, and the nodes in the `us-west-2` region are:
- 5x `us-west-2a`
- 5x `us-west-2b`
- 5x `us-west-2c`

The operator would deploy 3 cassandra nodes in each availability zone.

If the nodes were:
- 1x `us-west-2a`
- 10x `us-west-2b`
- 15x `us-west-2c`

Then the cassandra node distribution would probably end up similar to this:
- 1x `us-west-2a`
- 4x `us-west-2b`
- 4x `us-west-2c`

## Other parameters

### Endpoint Snitch

To let cassandra know about the topology, a different Snitch needs to be set:
```
ENDPOINT_SNITCH=GossipingPropertyFileSnitch
```
The GossipingPropertyFileSnitch lets cassandra read the datacenter and rack information from a local file which the operator generates from
the `NODE_TOPOLOGY`.

### Node Anti-Affinity

This prefents the cluster to schedule two cassandra nodes on to the same Kubernetes node.
```
NODE_ANTI_AFFINITY=true
```

If this feature is enabled, you *must* have at least that many Kubernetes nodes in your cluster as you use in the NODE_TOPOLOGY definition.

### Full list of required parameters

```
    ENDPOINT_SNITCH=GossipingPropertyFileSnitch
    NODE_ANTI_AFFINITY=true
    NODE_TOPOLOGY=<the cluster topology>
```