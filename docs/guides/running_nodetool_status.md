# Running `nodetool status`

## Introduction

The KUDO Cassandra operator runs Apache Cassandra node containers.

Subsequent commands require the KUDO Cassandra instance name and the Kubernetes
namespace it's running on, via the `instance_name` and `namespace_name`
variables respectively.

```bash
instance_name="cassandra"
namespace_name="production"
```

## Requirements

- You have a running KUDO Cassandra instance
- You have permission to run `kubectl exec` in KUDO Cassandra containers, or you
  can run a container in the Kubernetes cluster

## Steps

### Running `nodetool status` on a Cassandra cluster without SSL encryption

```bash
pod="0"

kubectl exec "${instance_name}-node-${pod}" \
        -n "${instance_namespace}" \
        -c cassandra \
        -- \
        bash -c "nodetool status"
```
