#!/usr/bin/env bash
# shellcheck disable=SC2039

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}/..")"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

readonly DEBUG="${DEBUG:=false}"
readonly VERBOSE="${VERBOSE:=false}"
CHECK_ONLY=false

if [ "${DEBUG}" == "true" ]; then
  set -x
fi

if [[ $# > 0 ]]; then
  case $1 in
  --check-only)
    CHECK_ONLY="true"
    ;;
  *)
    echo "Usage: $0 [--check-only]" >&2
    exit 1
    ;;
  esac
fi

log () {
  if [ "${VERBOSE}" == "true" ]; then
    echo "${*}"
  fi
}

declare -a env_vars
mapfile -t env_vars < <(sed -E \
                            's/export[[:space:]]+([[:alnum:]_]+)=.*/\1/g' \
                            "${project_directory}/metadata.sh" )

declare -a templates
mapfile -t templates < <(find "${project_directory}/templates" -type f)

for i in "${!env_vars[@]}" ; do
  env_vars[$i]="\${${env_vars[$i]}}"
done

join () { local IFS="${1}"; shift; echo "${*}"; }

readonly ENV_VARS_STRING="$(join , "${env_vars[@]}")"

ret=0

for template in "${templates[@]}"; do
  output_file_directory="$(dirname "${template}" | sed -e "s|${project_directory}/templates|${project_directory}|")"
  output_file="${output_file_directory}/$(basename "${template}" .template)"
  if [[ ${CHECK_ONLY} == true ]]; then
    echo "Checking that '${template}' matches '${output_file}'..." >&2
    envsubst "${ENV_VARS_STRING}" < "${template}" | diff -u "${output_file}" - >&2
    check_ret=$?
    if [[ ${check_ret} -ne 0 ]]; then
      ret=1
    else
      echo "OK" >&2
    fi
  else
    log "compiling '${template}' to '${output_file}'"
    envsubst "${ENV_VARS_STRING}" < "${template}" > "${output_file}"
  fi
done
exit ${ret}
