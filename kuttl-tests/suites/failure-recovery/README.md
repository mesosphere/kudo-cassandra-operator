This test ensures that the recovery controller works correctly

Things to keep in mind:

- Before a k8s node can be killed, all C\* nodes have to be in UN state, UJ is
  not enough, especially with only 2 nodes Otherwise the restarting node-0 can't
  find the old IP in the gossip

Failed approaches:

- Use `docker kill <node-with-cassandra-pod>` This fails because kind blocks on
  "collecting cluster logs" if a kind-worker is killed. Might be a bug in kind

- Use `node drain <node-with-cassandra-pod>` Drain seems not to be able to
  drain/evict the stateful set pod

- Use `kubectl delete node <node-with-cassandra-pod>` This more or less works,
  the problem is that sometimes the old Cassandra-Process keeps running. This
  leads to an error on the replacement node, as the replacement node can connect
  to the old IP and fails with "can not replace live node"

  - One workaround for this was to try to use the bootstrap on the new node to
    connect to the old node and drain/shut it down This failed in every case,
    although Cassandra itself was easily connecting to the old node. Maybe
    something about JMX/RPC
