#!/usr/bin/env bash

readonly KUDO_CASSANDRA_OPERATOR_NAME="${KUDO_CASSANDRA_OPERATOR_NAME:-cassandra}"
readonly KUDO_CASSANDRA_VERSION="${KUDO_CASSANDRA_VERSION:-0.1.0}"
readonly KUDO_CASSANDRA_INSTANCE_NAME="${KUDO_CASSANDRA_INSTANCE_NAME:-cassandra}"
readonly KUDO_CASSANDRA_INSTANCE_NAMESPACE="${KUDO_CASSANDRA_INSTANCE_NAMESPACE:-kudo-cassandra}"

kubectl delete instance \
        "${KUDO_CASSANDRA_INSTANCE_NAME}" \
        -n "${KUDO_CASSANDRA_INSTANCE_NAMESPACE}"

kubectl delete operatorversion \
        "${KUDO_CASSANDRA_OPERATOR_NAME}-${KUDO_CASSANDRA_VERSION}" \
        -n "${KUDO_CASSANDRA_INSTANCE_NAMESPACE}"

kubectl delete operator \
        "${KUDO_CASSANDRA_OPERATOR_NAME}" \
        -n "${KUDO_CASSANDRA_INSTANCE_NAMESPACE}"

declare -a PVCS
mapfile -t PVCS < <(
  kubectl get pvc \
          -n "${KUDO_CASSANDRA_INSTANCE_NAMESPACE}" \
          -o 'jsonpath={.items[*].metadata.name}' \
    | tr ' ' '\n'
)

for pvc in "${PVCS[@]}"; do
  kubectl delete "pvc/${pvc}" -n "${KUDO_CASSANDRA_INSTANCE_NAMESPACE}"
done
