#!/usr/bin/env bash

# This script contains metadata that is either used in other scripts or expanded
# into templates via `tools/compile_templates.sh`.

# "Shadowing" these two environment variables so that they don't affect
# similarly named environment variables in other scripts loading this script.
_script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
_project_directory="$(readlink -f "${_script_directory}")"

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

export OPERATOR_DIRECTORY="${_project_directory}/operator"
export VENDOR_DIRECTORY="${_project_directory}/shared/vendor"

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
############################## Docker images ###################################
################################################################################

export CASSANDRA_DOCKER_IMAGE_FROM="cassandra:${CASSANDRA_VERSION}"
export CASSANDRA_DOCKER_IMAGE_NAMESPACE="mpereira"
export CASSANDRA_DOCKER_IMAGE_NAME="${OPERATOR_NAME}"
export CASSANDRA_DOCKER_IMAGE_TAG="${OPERATOR_VERSION}-${CASSANDRA_VERSION}"
export CASSANDRA_DOCKER_IMAGE="${CASSANDRA_DOCKER_IMAGE_NAMESPACE}/${CASSANDRA_DOCKER_IMAGE_NAME}:${CASSANDRA_DOCKER_IMAGE_TAG}"

export PROMETHEUS_EXPORTER_DOCKER_IMAGE_FROM="criteord/cassandra_exporter:2.2.1"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAMESPACE="viivek46"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAME="prometheus-exporter"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE_TAG="v1.0"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE="${PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAMESPACE}/${PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAME}:${PROMETHEUS_EXPORTER_DOCKER_IMAGE_TAG}"

################################################################################
################################# Testing ######################################
################################################################################

export INTEGRATION_TESTS_DOCKER_IMAGE="golang:1.13.1-stretch"
