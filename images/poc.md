
## Cassandra Failover Demo

![failover](./resources/poc-diagram.png) 



### Setup

Checkout the KUDO Cassandra github repo
```
git clone https://github.com/mesosphere/kudo-cassandra-operator.git 
cd kudo-cassandra-operator
git checkout node-replace-poc
```

Install the KUDO Cassandra operator

```
kubectl kudo install ./operator --parameter-file=./images/topology/aws-topology.yaml
```

Watch the cassandra pods till we have the cluster up and healthy

```
kubectl get pods -w
```

Install the workload

```
kubectl apply -f ./images/manifests/
```

Get the benchmark endpoint

```
http://konvoy-cluster-url/verizon/poc/
```

Start with small test against the host `http://shopping-app-service` and check the shopping cart with 

```
kubectl exec -ti cassandra-instance-ring1-node-0  -- cqlsh -e "SELECT * FROM shopping.carts"
```

### Delete the whole ring1

```
kubectl delete pod cassandra-instance-ring1-node-0 cassandra-instance-ring1-node-1 cassandra-instance-ring1-node-2
```

