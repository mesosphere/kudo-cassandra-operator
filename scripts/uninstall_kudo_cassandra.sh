#!/usr/bin/env bash

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly PROJECT_DIRECTORY="$(readlink -f "${SCRIPT_DIRECTORY}/..")"

# shellcheck source=../metadata.sh
source "${PROJECT_DIRECTORY}/metadata.sh"

kudo_cassandra_version=
kudo_cassandra_instance_name=
kudo_cassandra_instance_namespace=

while [[ ${#} -gt 0 ]]; do
  # TODO(mpereira): handle parameters passed in as "parameter=value";
  parameter="${1}"

  case "${parameter}" in
    --version|-v)
      kudo_cassandra_version="${2}"
      shift
      ;;
    --instance|-i)
      kudo_cassandra_instance_name="${2}"
      shift
      ;;
    --namespace|-n)
      kudo_cassandra_instance_namespace="${2}"
      shift
      ;;
    *)
      ;;
  esac

  shift
done

kudo_cassandra_version="${kudo_cassandra_version:-0.1.0}"
kudo_cassandra_instance_name="${kudo_cassandra_instance_name:-cassandra}"
kudo_cassandra_instance_namespace="${kudo_cassandra_instance_namespace:-kudo-cassandra}"

kubectl delete instance \
        "${kudo_cassandra_instance_name}" \
        -n "${kudo_cassandra_instance_namespace}"

kubectl delete operatorversion \
        "cassandra-${kudo_cassandra_version}" \
        -n "${kudo_cassandra_instance_namespace}"

kubectl delete operator \
        "cassandra" \
        -n "${kudo_cassandra_instance_namespace}"

declare -a PVCS
mapfile -t PVCS < <(
  kubectl get pvc \
          -n "${kudo_cassandra_instance_namespace}" \
          -o 'jsonpath={.items[*].metadata.name}' \
    | tr ' ' '\n'
)

for pvc in "${PVCS[@]}"; do
  kubectl delete "pvc/${pvc}" -n "${kudo_cassandra_instance_namespace}"
done
