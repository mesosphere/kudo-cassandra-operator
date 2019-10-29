#!/usr/bin/env bash
# shellcheck disable=SC2039

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}/..")"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

readonly DEBUG="${DEBUG:=false}"
readonly VERBOSE="${VERBOSE:=false}"

if [ "${DEBUG}" == "true" ]; then
  set -x
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

for template in "${templates[@]}"; do
  output_file_directory="$(dirname "${template}" | sed -e "s|${project_directory}/templates|${project_directory}|")"
  output_file="${output_file_directory}/$(basename "${template}" .template)"
  log "compiling '${template}' to '${output_file}'"
  envsubst "${ENV_VARS_STRING}" < "${template}" > "${output_file}"
done
