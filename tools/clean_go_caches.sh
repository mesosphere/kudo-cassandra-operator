#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

go clean -cache -testcache -modcache
