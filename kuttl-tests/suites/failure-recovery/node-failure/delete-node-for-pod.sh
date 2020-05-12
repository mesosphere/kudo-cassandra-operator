#!/bin/bash

POD_NAME=$1
# $2 is "--namespace"
NAMESPACE=$3

NODE_NAME=`kubectl get pod ${POD_NAME} -n ${NAMESPACE} -o=custom-columns=NODE:.spec.nodeName --no-headers`

echo "Delete node $NODE_NAME"

kubectl delete node ${NODE_NAME}