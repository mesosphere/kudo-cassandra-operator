#!/bin/bash

POD_NAME=$1
# $2 is "--namespace"
NAMESPACE=$3

NODE_NAME=`kubectl get pod ${POD_NAME} -n ${NAMESPACE} -o=custom-columns=NODE:.spec.nodeName --no-headers`

echo "Cordon node $NODE_NAME"
kubectl cordon ${NODE_NAME}

echo "Deleting pod {$POD_NAME}"
kubectl delete pod ${POD_NAME} -n ${NAMESPACE}

echo "Delete node $NODE_NAME"
kubectl delete node ${NODE_NAME}