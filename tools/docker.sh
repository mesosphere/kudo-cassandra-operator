#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"

if [[ $# -eq 0 ]] ; then
    echo 'script requires the command to run in dockered linux env' >&2
    echo 'example:  ./tools/docker.sh ./tools/generate_parameters_markdown.py ' >&2
    exit 0
fi

docker build "${REPO_ROOT}/tools" -t kudo-cassandra-tools


# Run docker and copy all ENV vars except any that contain PATH
docker run --rm -u "$(id -u):$(id -g)" --env-file <(env | grep -v "PATH.*=") -v "${REPO_ROOT}:${REPO_ROOT}" -w "$(pwd)" kudo-cassandra-tools "$@"
