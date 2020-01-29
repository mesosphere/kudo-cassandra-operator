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

# More details about KUDO Versioning:
# https://github.com/kudobuilder/kudo/pull/1028
export OPERATOR_VERSION="0.1.2"

# This should be an empty string on stable branches and "-SNAPSHOT" on
# non-stable branches.
export POSSIBLE_SNAPSHOT_SUFFIX="-SNAPSHOT"

export OPERATOR_DIRECTORY="${_project_directory}/operator"
export VENDOR_DIRECTORY="${_project_directory}/shared/vendor"

################################################################################
############################### Dependencies ###################################
################################################################################

# http://www.apache.org/dyn/closer.lua/cassandra/3.11.5
# https://hub.docker.com/_/cassandra
# https://github.com/docker-library/cassandra/blob/master/3.11/Dockerfile
export CASSANDRA_VERSION="3.11.5"

# https://github.com/kudobuilder/kudo/releases/tag/v0.10.1
export KUDO_VERSION="0.10.1"

export KUBERNETES_VERSION="1.15.0"

export CASSANDRA_EXPORTER_DOCKER_IMAGE="criteord/cassandra_exporter"
export CASSANDRA_EXPORTER_VERSION="2.2.1"

################################################################################
############################## Docker images ###################################
################################################################################

export CASSANDRA_DOCKER_IMAGE_FROM="cassandra:${CASSANDRA_VERSION}"
export CASSANDRA_DOCKER_IMAGE_NAMESPACE="mesosphere"
export CASSANDRA_DOCKER_IMAGE_NAME="${OPERATOR_NAME}"
export CASSANDRA_DOCKER_IMAGE_TAG="${CASSANDRA_VERSION}-${OPERATOR_VERSION}${POSSIBLE_SNAPSHOT_SUFFIX}"
export CASSANDRA_DOCKER_IMAGE="${CASSANDRA_DOCKER_IMAGE_NAMESPACE}/${CASSANDRA_DOCKER_IMAGE_NAME}:${CASSANDRA_DOCKER_IMAGE_TAG}"

export PROMETHEUS_EXPORTER_DOCKER_IMAGE_FROM="${CASSANDRA_EXPORTER_DOCKER_IMAGE}:${CASSANDRA_EXPORTER_VERSION}"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAMESPACE="mesosphere"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAME="cassandra-prometheus-exporter"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE_TAG="${CASSANDRA_EXPORTER_VERSION}-${OPERATOR_VERSION}${POSSIBLE_SNAPSHOT_SUFFIX}"
export PROMETHEUS_EXPORTER_DOCKER_IMAGE="${PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAMESPACE}/${PROMETHEUS_EXPORTER_DOCKER_IMAGE_NAME}:${PROMETHEUS_EXPORTER_DOCKER_IMAGE_TAG}"

################################################################################
################################# Testing ######################################
################################################################################

export INTEGRATION_TESTS_DOCKER_IMAGE="golang:1.13.1-stretch"
