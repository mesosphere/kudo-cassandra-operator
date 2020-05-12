#!/bin/bash

NODE_NAME=$1
EXPECTED_NODE_COUNT=$2
# $3 is "--namespace"
NAMESPACE=$4

# Fetch the list of current nodes from cassandra and parse lines with UN/DN/... etc from the output
ACTUAL_NODE_COUNT=`kubectl exec -n ${NAMESPACE} ${NODE_NAME} nodetool status | grep -E "^\w{2}\s+.*$" | wc -l`

echo "Expected Node count: $EXPECTED_NODE_COUNT, Actual Node count: $ACTUAL_NODE_COUNT"

# Return corresponding exit code
[[ ${ACTUAL_NODE_COUNT}  == ${EXPECTED_NODE_COUNT} ]]