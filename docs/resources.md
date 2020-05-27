# Resources

## Computable Resources

KUDO Cassandra by default requests 1.5 cpu and 4.5Gi of memory for each
Cassandra node. The limits by default are 2 cpu and 4.5Gi of memory.

Those requests and limits can be tuned using the parameters:

- NODE_CPU_MC
- NODE_CPU_LIMIT_MC
- NODE_MEM_MIB
- NODE_MEM_LIMIT_MIB
- PROMETHEUS_EXPORTER_CPU_MC
- PROMETHEUS_EXPORTER_CPU_LIMIT_MC
- PROMETHEUS_EXPORTER_MEM_MIB
- PROMETHEUS_EXPORTER_MEM_LIMIT_MIB
- RECOVERY_CONTROLLER_CPU_MC
- RECOVERY_CONTROLLER_CPU_LIMIT_MC
- RECOVERY_CONTROLLER_MEM_MIB
- RECOVERY_CONTROLLER_MEM_LIMIT_MIB

## Storage resources

By default KUDO Cassandra uses 20GiB PV. This isn't recommended for production
use. Please refer to [production](./production.md) docs to see the storage and
computer resources recommendations

#### Cassandra container

```
resources:
  limits:
    cpu: 1
    memory: 4Gi
  requests:
    cpu: 1
    memory: 4Gi
```

#### Bootstrap init container

```
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

#### prometheus exporter sidecar

```
resources:
  limits:
    cpu: 1
    memory: 512Mi
  requests:
    cpu: 500m
    memory: 512Mi
```

#### cassandra-recovery controller pod

```
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 50Mi
```

## Kubernetes Objects

KUDO Cassandra is delivered with a different set of features, enabling or
disabling those features creates more or fewer objects in Kubernetes.

Let’s take a look at the resources created by default when doing a simple

```
$ kubectl kudo install cassandra
operator.kudo.dev/v1beta1/cassandra created
operatorversion.kudo.dev/v1beta1/cassandra-<version> created
instance.kudo.dev/v1beta1/cassandra-instance created
instance.kudo.dev/v1beta1/cassandra-instance ready

$ kubectl tree instance cassandra-instance
NAMESPACE  NAME                                                             READY  REASON  AGE
default    Instance/cassandra-instance                                      -              6m11s
default    ├─ConfigMap/cassandra-instance-cassandra-env-sh                  -              6m9s
default    ├─ConfigMap/cassandra-instance-cassandra-exporter-config-yml     -              6m9s
default    ├─ConfigMap/cassandra-instance-generate-cassandra-yaml           -              6m9s
default    ├─ConfigMap/cassandra-instance-generate-cqlshrc-sh               -              6m9s
default    ├─ConfigMap/cassandra-instance-generate-nodetool-ssl-properties  -              6m9s
default    ├─ConfigMap/cassandra-instance-generate-tls-artifacts-sh         -              6m9s
default    ├─ConfigMap/cassandra-instance-jvm-options                       -              6m9s
default    ├─ConfigMap/cassandra-instance-node-scripts                      -              6m9s
default    ├─ConfigMap/cassandra-instance-topology-lock                     -              6m9s
default    ├─PodDisruptionBudget/cassandra-instance-pdb                     -              6m9s
default    ├─Role/cassandra-instance-node-role                              -              6m10s
default    ├─Role/cassandra-instance-role                                   -              6m9s
default    ├─RoleBinding/cassandra-instance-binding                         -              6m9s
default    ├─RoleBinding/cassandra-instance-node-default-binding            -              6m10s
default    ├─Secret/cassandra-instance-tls-store-credentials                -              6m9s
default    ├─Service/cassandra-instance-svc                                 -              6m9s
default    ├─ServiceAccount/cassandra-instance-sa                           -              6m9s
default    ├─ServiceMonitor/cassandra-instance-monitor                      -              6m9s
default    └─StatefulSet/cassandra-instance-node                            -              6m9s
default      ├─ControllerRevision/cassandra-instance-node-659c89769d        -              2m24s
default      ├─Pod/cassandra-instance-node-0                                True           2m14s
default      ├─Pod/cassandra-instance-node-1                                True           94s
default      └─Pod/cassandra-instance-node-2                                True           53s
```

### Statefulset

Statefulsets are designed to manage stateful workload in Kubernetes. KUDO
Cassandra uses statefulsets. The operator by default uses `OrderedReady`pod
management policy. Which guarantees that pods are created sequentially. This
makes sure that when the Cassandra cluster is coming up, only one node starts at
the same time. Pod names are <instance-name>-node-<ordinal-index> starting from
ordinal-index 0. For example a 3 node cluster created using KUDO Cassandra
instance name cass-prod will have these pods:

```
$ kubectl get pods
NAME               READY   STATUS    RESTARTS   AGE
cass-prod-node-0   1/1     Running   0          101s
cass-prod-node-1   1/1     Running   0          49s
cass-prod-node-2   1/1     Running   0          8s
```

### Configmaps

KUDO Cassandra generates the configurable scripts and properties used in KUDO
Cassandra operator as configmaps objects PodDisruptionBudget KUDO Cassandra
limits the number of pods that are down simultaneously. For Cassandra’s service
to work without interruptions, especially when the quorum-based applications are
running on top of Cassandra we would like to guarantee that the number of
replicas running is never brought below the number required for a quorum, even
temporarily. Unlike a regular pod deletion, for the KUDO Cassandra pod eviction
the API server may reject operation if the eviction would cause a
disruption budget to be disrupted.

### ServiceAccount / Role / RoleBinding

KUDO Cassandra creates one service account which is attached with two types of
roles. The service account is <instance-name>-sa and then the Roles are:

```
<instance-name>-node-role
<instance-name>-role
```

The node role is used by KUDO Cassandra recovery feature to use kubernetes API
to detect any deleted kubelets. And also to free the KUDO Cassandra PVC Claim
Ref if that is necessary. The generic role, `<instance-name>-role` is used by
the Cassandra node bootstrap binary to update the topology-lock configmap so it
has access to the configmaps

### Secrets

KUDO Cassandra creates the TLS store credentials as a secret
<instance-name>-tls-store-credentials. Those credentials are used as
keystore/truststore credentials, when adding certificates to them. Which is done
when SSL feature is enabled for KUDO Cassandra.

### Service

KUDO Cassandra creates a headless service <instance-name>-svc (ClusterIP type).
This service is used for internal access from inside the kubernetes cluster.

```
$ kubectl get svc
NAME                     TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                               AGE
cassandra-instance-svc   ClusterIP   10.0.51.69   <none>        7000/TCP,7001/TCP,9042/TCP,7200/TCP   26m
```

The service exposes the storage port on 7000 by default and if SSL is enabled
then on 7001. The native transport port on 9042 by default and metrics are
exposed on the port 7200 by default.

`RPC` port is disabled by default and can be enabled using the parameter
START_RPC=true, which will expose the RPC port on 9160 by default. All above
ports can be exposed on custom ports using the parameters:

```
STORAGE_PORT
SSL_STORAGE_PORT
NATIVE_TRANSPORT_PORT
RPC_PORT
```

### ServiceMonitor

KUDO Cassandra integrates with prometheus-operator by default. The
ServiceMonitor uses the label kudo.dev/servicemonitor and kudo.dev/instance
labels to discover the Cassandra pods.

To disable the monitoring user can start the KUDO Cassandra with parameter
`PROMETHEUS_EXPORTER_ENABLED=false`. Read more information about monitoring in
the Monitoring section
