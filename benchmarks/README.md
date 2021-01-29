# Benchmarks

## Running a benchmark session

### Set up some KUDO Cassandra instance information

1. Ensure a configuration with clsuter
2. cd mwt
3. read [README](mwt/README.md)

```bash
kudo_cassandra_operator_name="cassandra"
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

### Deploy `cassandra-stress` operator

This will deploy the `cassandra-stress` operator based on the parameters in `workloads/stress-params.yaml` file.

```bash
k kuttl test workload/
```

### Check workload pods

```bash
kubectl get po -n "${kudo_cassandra_instance_namespace}"
```

### Delete workload

```bash
kubectl kudo uninstall --instance=cassandra-stress -n "${kudo_cassandra_instance_namespace}"
```

### Uninstall Cassandra operator

From `mwt`  
Run `kubectl kuttl test teardown/`
