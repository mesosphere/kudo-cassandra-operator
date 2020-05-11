#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

go get github.com/a8m/envsubst/cmd/envsubst
source "$(dirname "$0")/../shared/teamcity/internal/common.sh"
parse_options "" "$@"

export IMAGE_DISAMBIGUATION_SUFFIX="$(get_image_disambiguation_suffix)"
source "$(dirname "$0")/../metadata.sh"

envsubst < "templates/00-install.template.yaml" > "tests/install-test/00-install.yaml"
env
