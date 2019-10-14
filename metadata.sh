#!/usr/bin/env bash

# This script contains metadata that is either used in other scripts or expanded
# into templates via `tools/compile_templates.sh`.

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
project_directory="$(readlink -f "${script_directory}")"

################################################################################
################################# Operator #####################################
################################################################################

# https://github.com/mesosphere/kudo-cassandra-operator
export PROJECT_NAME="kudo-cassandra-operator"
export OPERATOR_NAME="cassandra"
# KUDO still doesn't support snapshots, or compound versions yet. Check out:
# - https://github.com/kudobuilder/kudo/pull/889
# - https://github.com/kudobuilder/kudo/issues/163
export OPERATOR_VERSION="0.1.0"
export OPERATOR_DIRECTORY="${project_directory}/operator"
export VENDOR_DIRECTORY="${project_directory}/shared/vendor"

################################################################################
############################### Dependencies ###################################
################################################################################

# http://www.apache.org/dyn/closer.lua/cassandra/3.11.4
# https://hub.docker.com/_/cassandra
# https://github.com/docker-library/cassandra/blob/master/3.11/Dockerfile
export CASSANDRA_VERSION="3.11.4"

# https://github.com/kudobuilder/kudo/releases/tag/v0.7.4
export KUDO_VERSION="0.7.4"

export KUBERNETES_VERSION="1.15.0"

################################################################################
################################# Testing ######################################
################################################################################

export INTEGRATION_TESTS_DOCKER_IMAGE="golang:1.13.1-stretch"
