#!/bin/bash

NODE_NAME=$1
EXPECTED_DOWN_NODE_COUNT=$2
# $3 is "--namespace"
NAMESPACE=$4

CMD='nodetool --ssl status'

for i in {1..15}; do
    # Fetch the list of current nodes from cassandra and parse lines with UN/DN/... etc from the output
    DOWN_NODE_COUNT=`kubectl exec -n ${NAMESPACE} ${NODE_NAME} -- ${CMD} | grep -E "^DN\s+.*$" | wc -l`

    echo "Expected Down Node count: $EXPECTED_DOWN_NODE_COUNT, Actual Node count: $DOWN_NODE_COUNT (Try $i)"

    if [[ ${EXPECTED_DOWN_NODE_COUNT} == ${DOWN_NODE_COUNT} ]]; then
        echo "Found matching down node count"
        exit 0
    fi

    NODE_IPS=`kubectl exec -n ${NAMESPACE} ${NODE_NAME} -- ${CMD} | sed -nE 's/^UN[[:space:]]+([0-9.]+).*/\1/p'`
    echo "Node IPS: $NODE_IPS"
    for IP in ${NODE_IPS}; do
        echo "==> Try Status on $IP from node-0 bootstrap"
        kubectl exec -n ${NAMESPACE} cassandra-node-0 --container bootstrap -- nodetool --ssl --host ${IP} --port 7199 status
        echo "==> Try Status on $IP from node-0 cassandra"
        kubectl exec -n ${NAMESPACE} cassandra-node-0 --container bootstrap -- nodetool --ssl --host ${IP} --port 7199 status
        echo "==> Try Status on $IP from node-1 cassandra"
        kubectl exec -n ${NAMESPACE} cassandra-node-1 --container bootstrap -- nodetool --ssl --host ${IP} --port 7199 status
    done

    sleep 5
done

# Return corresponding exit code
exit 1