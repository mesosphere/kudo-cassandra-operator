### KUDO Cassandra Bootstrap

The KUDO Cassandra bootstrap binary fills in the missing capabilities we need to run a production grade Cassandra in Kubernetes.


#### Cluster IP topology

For each bootstrap, it updates the pod ip, so we don't need the sticky IP addresses. 
And in case the node isn't bootstrapped and a new IP is assigned, it makes sure to bootstrap the node with right flags to replace the old node.

#### Steps

1. get the pod IP in the CM
1. in case there is no IP updates the CM
1. in case there is an old IP and node is already bootstrapped updates the CM 
1. in case there is an old IP and node is not bootstrapped
    1. writes the old ip in the file `/var/lib/cassandra/replace.ip`
    1. waits for the node to be in UJ/UN state and updates the current IP in the CM
    1. clears the file `/var/lib/cassandra/replace.ip` for any next bootstrap