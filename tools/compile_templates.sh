#!/usr/bin/env bash
# shellcheck disable=SC2039

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly PROJECT_DIRECTORY="$(readlink -f "${SCRIPT_DIRECTORY}/..")"

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
                            "${PROJECT_DIRECTORY}/metadata.sh" )

declare -a templates
mapfile -t templates < <(find "${PROJECT_DIRECTORY}/templates" -type f)

# shellcheck source=../metadata.sh
source "${PROJECT_DIRECTORY}/metadata.sh"

for i in "${!env_vars[@]}" ; do
  env_vars[$i]="\${${env_vars[$i]}}"
done

join () { local IFS="${1}"; shift; echo "${*}"; }

readonly ENV_VARS_STRING="$(join , "${env_vars[@]}")"

for template in "${templates[@]}"; do
  output_file_directory="$(dirname "${template}" | sed -e "s|${PROJECT_DIRECTORY}/templates|${PROJECT_DIRECTORY}|")"
  output_file="${output_file_directory}/$(basename "${template}" .template)"
  log "compiling '${template}' to '${output_file}'"
  envsubst "${ENV_VARS_STRING}" < "${template}" > "${output_file}"
done
