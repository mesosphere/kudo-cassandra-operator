#!/usr/bin/env bash

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(readlink -f "${script_directory}")"

# We need to ignore the suffix for the purpose of checking templates.
IMAGE_DISAMBIGUATION_SUFFIX="" "${project_directory}/tools/compile_templates.sh" --check-only

"${project_directory}/tools/docker.sh" ./tools/generate_parameters_markdown.py

"${project_directory}/tools/docker.sh" ./tools/format_files.sh

if [ -n "$(git status --porcelain)" ]; then
  echo "Changes found after running one of the previous steps." >&2
  echo "Please make sure you follow the instructions in .github/pull_request_template.md" >&2
  echo "before sending a pull request." >&2
  git status --porcelain
  exit 1
fi
