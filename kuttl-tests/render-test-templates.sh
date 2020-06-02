#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

go get github.com/schneidexe/tpl
go build -o bin/tpl github.com/schneidexe/tpl
source "$(dirname "$0")/../shared/teamcity/internal/common.sh"
parse_options "" "$@"

export IMAGE_DISAMBIGUATION_SUFFIX="$(get_image_disambiguation_suffix)"
source "$(dirname "$0")/../metadata.sh"

env

for FILE_TEMPLATE in `find . -name "*.template"`; do
    FILE_NAME="${FILE_TEMPLATE%.template}"
    echo "Render Template $FILE_TEMPLATE -> $FILE_NAME";
    ./bin/tpl -t ${FILE_TEMPLATE} > ${FILE_NAME}
done
