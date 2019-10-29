#!/usr/bin/env bash
# shellcheck disable=SC2039

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}/..")"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

readonly kubectl="${KUBECTL_PATH:-kubectl}"

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
operator_instance_name="${operator_instance_name:-${OPERATOR_INSTANCE_NAME}}"
operator_instance_namespace="${operator_instance_namespace:-${OPERATOR_INSTANCE_NAMESPACE}}"

for parameter in operator_name \
                   operator_version \
                   operator_instance_name \
                   operator_instance_namespace; do
  if [[ -z ${!parameter} ]]; then
    echo "--${parameter} parameter is required" >&2
    exit 1
  fi
done

${kubectl} delete instance \
           "${operator_instance_name}" \
           -n "${operator_instance_namespace}"

# TODO(mpereira): add a flag to skip operatorversion deletion?
${kubectl} delete operatorversion \
           "${operator_name}-${operator_version}" \
           -n "${operator_instance_namespace}"

# TODO(mpereira): add a flag to skip operator deletion?
${kubectl} delete operator \
        "${operator_name}" \
        -n "${operator_instance_namespace}"

# TODO(mpereira): add a flag to skip pvc deletion?
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
