#!/usr/bin/env bash
# shellcheck disable=SC2039

readonly DEBUG="${DEBUG:=false}"
readonly COLOR="${COLOR:=true}"

if [ "${DEBUG}" == "true" ]; then
  set -x
fi

# YAML #########################################################################

# FIXME(mpereira): can't use Prettier for now since it doesn't support templated
# YAML. Is there something else we could use?
# Also, check out https://github.com/kudobuilder/kudo/issues/887.

# declare -a prettier_options

# if [ "${COLOR}" == "false" ]; then
#   prettier_options+=("--no-color")
# fi

# mapfile -t yaml_files < <(git ls-files -- ':!:shared' | grep -E '.ya?ml')
# prettier --loglevel warn --check "${prettier_options[@]}" "${yaml_files[@]}"

# readonly yaml_exit_code=$?

# Shell scripts ################################################################

declare -a shellcheck_options

if [ "${COLOR}" == "false" ]; then
  shellcheck_options+=("--color=never")
fi

# Get all files with a shebang.
mapfile -t shell_scripts < <(git ls-files -- ':!:shared' | xargs -I{} grep -El '^#!.+(sh|bash)' {})
shellcheck -ax -e SC1091 "${shellcheck_options[@]}" "${shell_scripts[@]}"

readonly shell_scripts_exit_code=$?

################################################################################

! ((shell_scripts_exit_code))
