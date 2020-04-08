#!/bin/bash

TOOL_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [[ $# -eq 0 ]] ; then
    echo 'script requires the command to run in dockered linux env'
    echo 'example:  ./tools/docker.sh ./tools/generate_parameters_markdown.py '
    exit 0
fi

cd $TOOL_DIR
docker build . -t cass-tools

cd $TOOL_DIR/..
docker run  -v $PWD:/opt/kudo-cassandra-operator cass-tools $1
