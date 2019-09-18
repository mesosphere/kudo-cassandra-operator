#!/usr/bin/env bash

readonly KUDO_CASSANDRA_VERSION="${KUDO_CASSANDRA_VERSION:-0.1.0}"
readonly KUBERNETES_NAMESPACE="${KUBERNETES_NAMESPACE:-kudo-cassandra}"

kubectl delete instance cassandra -n "${KUBERNETES_NAMESPACE}"
kubectl delete operatorversion "cassandra-${KUDO_CASSANDRA_VERSION}" -n "${KUBERNETES_NAMESPACE}"
kubectl delete operator cassandra -n "${KUBERNETES_NAMESPACE}"
# TODO(mpereira): get pvcs and iterate deleting each.
kubectl delete pvc/var-lib-cassandra-cassandra-cassandra-0 -n "${KUBERNETES_NAMESPACE}"
kubectl delete pvc/var-lib-cassandra-cassandra-cassandra-1 -n "${KUBERNETES_NAMESPACE}"
kubectl delete pvc/var-lib-cassandra-cassandra-cassandra-2 -n "${KUBERNETES_NAMESPACE}"
