# Evicting a KUDO Cassandra Node

Usually KUDO Cassandra schedules the Cassandra Cluster Nodes randomly to the Kubernetes nodes. It supports
Anti-Affinity - multiple Cassandra nodes should not run on the same Kubernetes Nodes.

After a node has started, it usually will not move around in the cluster. The Cassandra Nodes require a persistent
volume, which normally is a local volume which will only be available on the specific Kubernetes Node where the 
Cassandra instance was initially started. (There are options like EBS volumes that can move around, but the focus here 
is on local volumes)

In some cases a cluster operator wants to evict a Cassandra Node from a specific Kubernetes Node.

## Required setup

Have a KUDO Cassandra cluster running, with at least the following parameters:

```yaml
RECOVERY_CONTROLLER: "true"
```

## Steps

### Taint or Cordon old node
To prevent that the evicted Cassandra instance will be restarted on the same node again, add a taint or cordon the
old node where the instance was running

### Mark pod for eviction
Add a new label to the node to be evicted:

```bash
kubectl label pod cassandra-node-0 kudo-cassandra/evict=true
```

This will trigger the recovery controller to unlink the PV and remove the PVC, so the pod can be rescheduled to a different
Kubernetes node.

After a while the old pod will be terminated and rescheduled on a different Kubernetes node.