#!/usr/bin/env bash
# shellcheck disable=SC2039

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly PROJECT_DIRECTORY="$(readlink -f "${SCRIPT_DIRECTORY}/..")"

# shellcheck source=../metadata.sh
source "${PROJECT_DIRECTORY}/metadata.sh"

kubectl="${KUBECTL_PATH:-kubectl}"

operator_name=
operator_version=
operator_instance_name=
operator_instance_namespace=

while [[ ${#} -gt 0 ]]; do
  # TODO(mpereira): handle parameters passed in as "parameter=value";
  parameter="${1}"

  case "${parameter}" in
    --operator|-o)
      operator_name="${2}"
      shift
      ;;
    --version|-v)
      operator_version="${2}"
      shift
      ;;
    --instance|-i)
      operator_instance_name="${2}"
      shift
      ;;
    --namespace|-n)
      operator_instance_namespace="${2}"
      shift
      ;;
    *)
      ;;
  esac

  shift
done

operator_name="${operator_name:-${OPERATOR_NAME}}"
operator_version="${operator_version:-${OPERATOR_VERSION}}"

${kubectl} delete instance \
           "${operator_instance_name}" \
           -n "${operator_instance_namespace}"

${kubectl} delete operatorversion \
           "${operator_name}-${operator_version}" \
           -n "${operator_instance_namespace}"

${kubectl} delete operator \
        "${operator_name}" \
        -n "${operator_instance_namespace}"

declare -a PVCS
mapfile -t PVCS < <(
  ${kubectl} get pvc \
          -n "${operator_instance_namespace}" \
          -o 'jsonpath={.items[*].metadata.name}' \
    | tr ' ' '\n'
)

for pvc in "${PVCS[@]}"; do
  ${kubectl} delete "pvc/${pvc}" -n "${operator_instance_namespace}"
done
