#!/usr/bin/env bash
# shellcheck disable=SC2039

set -euxo pipefail

readonly script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly project_directory="$(realpath -L $script_directory/..)"

# shellcheck source=../metadata.sh
source "${project_directory}/metadata.sh"

if [[ -n ${IMAGE_DISAMBIGUATION_SUFFIX:-} ]]; then
  # Refresh templated files to pick up the suffix, if explicitly set.
  "${project_directory}/tools/docker.sh" \
    env "IMAGE_DISAMBIGUATION_SUFFIX=${IMAGE_DISAMBIGUATION_SUFFIX}" \
    "${project_directory}/tools/compile_templates.sh"
fi

readonly cassandra_docker_image="${CASSANDRA_DOCKER_IMAGE:-}"
readonly prometheus_exporter_docker_image="${PROMETHEUS_EXPORTER_DOCKER_IMAGE:-}"
readonly medusa_backup_docker_image="${MEDUSA_BACKUP_DOCKER_IMAGE:-}"
readonly integration_tests_docker_image="${INTEGRATION_TESTS_DOCKER_IMAGE:-}"
readonly recovery_controller_docker_image="${RECOVERY_CONTROLLER_DOCKER_IMAGE:-}"

if [[ -z ${cassandra_docker_image} ]]; then
  echo "Missing CASSANDRA_DOCKER_IMAGE" >&2
  exit 1
fi

if [[ -z ${prometheus_exporter_docker_image} ]]; then
  echo "Missing PROMETHEUS_EXPORTER_DOCKER_IMAGE" >&2
  exit 1
fi

if [[ -z ${medusa_backup_docker_image} ]]; then
  echo "Missing MEDUSA_BACKUP_DOCKER_IMAGE" >&2
  exit 1
fi

if [[ -z ${integration_tests_docker_image} ]]; then
  echo "Missing INTEGRATION_TESTS_DOCKER_IMAGE" >&2
  exit 1
fi

if [[ -z ${recovery_controller_docker_image} ]]; then
  echo "Missing RECOVERY_CONTROLLER_DOCKER_IMAGE" >&2
  exit 1
fi

docker image build \
       -t "${cassandra_docker_image}" \
       -f "${project_directory}/images/Dockerfile" \
       "${project_directory}/images"

docker image build \
       -t "${prometheus_exporter_docker_image}" \
       -f "${project_directory}/images/Dockerfile.prometheus-exporter" \
       "${project_directory}/images"

docker image build \
       -t "${medusa_backup_docker_image}" \
       -f "${project_directory}/images/Dockerfile.medusa-backup" \
       "${project_directory}/images"

docker image build \
       -t "${integration_tests_docker_image}" \
       -f "${project_directory}/images/Dockerfile.integration-tests" \
       "${project_directory}/images"

docker image build \
       -t "${recovery_controller_docker_image}" \
       -f "${project_directory}/images/Dockerfile.recovery-controller" \
       "${project_directory}/images"

if [[ "${1:-}" == "push" ]]; then
  docker push "${cassandra_docker_image}"
  docker push "${prometheus_exporter_docker_image}"
  docker push "${medusa_backup_docker_image}"
  docker push "${integration_tests_docker_image}"
  docker push "${recovery_controller_docker_image}"
fi

if [[ "${1:-}" == "kind-load" ]]; then
  kind load docker-image "${cassandra_docker_image}"
  kind load docker-image "${prometheus_exporter_docker_image}"
  kind load docker-image "${medusa_backup_docker_image}"
  kind load docker-image "${integration_tests_docker_image}"
  kind load docker-image "${recovery_controller_docker_image}"
fi
