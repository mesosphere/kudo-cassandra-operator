#!/usr/bin/env bash
# shellcheck disable=SC2039

# Dependencies:
# - prettier
# - goimports

# YAML #########################################################################

# FIXME(mpereira): can't use Prettier for now since it doesn't support templated
# YAML. Is there something else we could use?
# Also, check out https://github.com/kudobuilder/kudo/issues/887.

# declare -a prettier_options

# mapfile -t yaml_files < <(git ls-files -- ':!:shared' | grep -E '.ya?ml')
# prettier --write --no-bracket-spacing "${prettier_options[@]}" "${yaml_files[@]}"

# readonly yaml_exit_code=$?

# FIXME(mpereira): see FIXME above.
readonly yaml_exit_code=0

# Go ###########################################################################

goimports -l -w .

readonly go_exit_code=$?

################################################################################

! ((yaml_exit_code | go_exit_code))
