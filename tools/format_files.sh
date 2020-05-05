#!/usr/bin/env bash
# shellcheck disable=SC2039

# Dependencies:
# - prettier
# - goimports

declare -a global_prettier_options
cd "$(dirname "$0")/.."

# YAML #########################################################################

# FIXME(mpereira): can't use Prettier for now since it doesn't support templated
# YAML. Is there something else we could use?
# Also, check out https://github.com/kudobuilder/kudo/issues/887.

# mapfile -t yaml_files < <(git ls-files -- ':!:shared' | grep -E '\.ya?ml$')
# prettier --write --no-bracket-spacing "${global_prettier_options[@]}" "${yaml_files[@]}"

# readonly yaml_exit_code=$?

# FIXME(mpereira): see FIXME above.
readonly yaml_exit_code=0

# Go ###########################################################################

goimports -l -w .

readonly go_exit_code=$?

# Markdown #####################################################################

mapfile -t markdown_files < <(git ls-files -- ':!:shared' | grep -E '\.md$')
prettier --parser markdown \
         --prose-wrap always \
         --write \
         "${global_prettier_options[@]}" \
         "${markdown_files[@]}"

readonly markdown_exit_code=$?

################################################################################

! ((yaml_exit_code | go_exit_code | markdown_exit_code))
