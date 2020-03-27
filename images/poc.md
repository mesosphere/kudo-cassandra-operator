
## Steps


Start controller watching events
```
cd cassandra-recovery
KUBECONFIG=path/to/kubeconfig go run main.go
```

## Install KUDO Cassandra

```
kubectl kudo install /Users/zain/go/src/github.com/mesosphere/kudo-cassandra-operator/operator/ -p NODE_MEM_MIB=256 -p PROMETHEUS_EXPORTER_ENABLED=false -p NODE_DOCKER_IMAGE=zmalikshxil/cassandra:3.11.5
```

replace `/Users/zain/go/src/github.com/mesosphere/kudo-cassandra-operator/operator/` with your own path to operator


if you do any changes in cassandra-bootstrap binary, make sure to build the docker image and push it
and replace `NODE_DOCKER_IMAGE=zmalikshxil/cassandra:3.11.5` with your own image

```
docker build . -t zmalikshxil/cassandra:3.11.5
docker push zmalikshxil/cassandra:3.11.5
```

---
Delete any kubernetes node where a Cassandra node is running to verify the POC

