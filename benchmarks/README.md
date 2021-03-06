# Benchmarks

## Running a benchmark session

### Set up some KUDO Cassandra instance information

1. Ensure a configuration with clsuter
2. cd mwt
3. read [README](mwt/README.md)

```bash
kudo_cassandra_operator_name="cassandra"
kudo_cassandra_operator_version="1.0.0"
kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cassandra"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"
```

### Setup and Verify KUDO Cassandra

From `mwt`  
Run `kubectl kuttl test setup/ --parallel 1 --skip-delete`  
Which will:

1. verify setup
2. install cassandra
3. wait for deployment to finish
4. output nodetool status

### Run `kassandra-stress`

This will create a Kubernetes deployment resource. Tweak the parameters as
desired.

```bash
workload_name="cassandra-workload-1"
```

```bash
./benchmarks/kassandra-stress \
  --kubernetes-namespace "${kudo_cassandra_instance_namespace}" \
  --workload-name "${workload_name}" \
  --hosts "${svc_endpoint}" \
  --number-of-clients 20 \
  --keyspace-name "cassandra_stress" \
  --one-keyspace-per-client false \
  --duration "1h"
```

### Generate workload by applying the deployment generated by `kassandra-stress`

```bash
kubectl apply -f benchmarks/kassandra_stress_replica_set.yaml
```

### Check workload deployment

```bash
kubectl get deployment -n "${kudo_cassandra_instance_namespace}"
```

### Check workload pods

```bash
kubectl get deployment -n "${kudo_cassandra_instance_namespace}"
```

### Delete workload

```bash
kubectl delete <deployment> -n "${kudo_cassandra_instance_namespace}"
```

### Uninstall operator

From `mwt`  
Run `kubectl kuttl test teardown/`
