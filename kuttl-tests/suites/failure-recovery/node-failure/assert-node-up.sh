#!/bin/bash

NODE_NAME=$1
EXPECTED_NODE_COUNT=$2
# $3 is "--namespace"
NAMESPACE=$4

CMD='nodetool --ssl status'

for i in {1..15}; do
    # Fetch the list of current nodes from cassandra and parse lines with UN... etc from the output
    ACTUAL_NODE_COUNT=`kubectl exec -n ${NAMESPACE} ${NODE_NAME} -- ${CMD} | grep -E "^UN\s+.*$" | wc -l`

    echo "Expected UN node count: $EXPECTED_NODE_COUNT, Actual Node count: $ACTUAL_NODE_COUNT (Try $i)"

    if [[ ${ACTUAL_NODE_COUNT} == ${EXPECTED_NODE_COUNT} ]]; then
        echo "Found matching node count"
        exit 0
    fi
    sleep 5
done

# Return corresponding exit code
exit 1