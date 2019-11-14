#!/usr/bin/env python3

import argparse
import json
import logging
import os.path
import sys
import tempfile

from utils import (
    run,
    clone_repository,
    random_short_string,
    configure_git_user,
    create_pull_request,
)

PROGRAM_NAME = os.path.basename(__file__)

log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%d %H:%M:%SZ",
)


KUDOBUILDER_OPERATORS_REPOSITORY = "kudobuilder/operators"


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Open a PR bringing a KUDO Operator's files to the "
        + "kudobuilder/operators repository"
    )

    parser.add_argument(
        "--operator-repository",
        type=str,
        required=True,
        help="The KUDO Operator repository to bring the files from "
        + "(e.g., mesosphere/kudo-cassandra-operator)",
    )
    parser.add_argument(
        "--operator-name",
        type=str,
        required=True,
        help="The name of the KUDO operator (e.g., cassandra, kafka, spark)",
    )
    parser.add_argument(
        "--operator-git-tag",
        type=str,
        required=True,
        help="The git tag in the KUDO operator repository to bring the "
        + "files from. This is also the directory name in "
        + "kudobuilder/operators, e.g., repository/cassandra/$OPERATOR_GIT_TAG",
    )
    parser.add_argument(
        "--github-token",
        type=str,
        required=True,
        help="The GitHub token used for opening the PR against kudobuilder/operators",
    )
    parser.add_argument(
        "--git-commit-message",
        type=str,
        help="The git commit message in the kudobuilder/operators PR branch. "
        + "This will also be the PR title",
    )
    parser.add_argument(
        "--operators-base-branch",
        type=str,
        default="master",
        help="The kudobuilder/operators branch to open a PR against",
    )
    parser.add_argument(
        "--git-user",
        type=str,
        default="git",
        help="The git user for cloning and pushing via SSH",
    )
    parser.add_argument(
        "--debug",
        action="store_true",
        default=False,
        help="Show debug output from all operations performed",
    )

    args = parser.parse_args()
    operator_repository = args.operator_repository
    operator_name = args.operator_name
    operator_git_tag = args.operator_git_tag
    github_token = args.github_token
    git_commit_message = args.git_commit_message or (
        f"Release {operator_name} {operator_git_tag} (automated commit)."
    )
    operators_base_branch = args.operators_base_branch
    git_user = args.git_user
    debug = args.debug

    base_directory = tempfile.mkdtemp("_kudo_operator_tools")
    operators_repository = KUDOBUILDER_OPERATORS_REPOSITORY
    operators_repository_url = (
        f"{git_user}@github.com:{operators_repository}.git"
    )
    operators_branch = (
        f"{operator_name}_{operator_git_tag}_{random_short_string()}"
    )

    rc, operators_directory, error_message = clone_repository(
        operators_repository_url, operators_base_branch, base_directory, debug
    )
    if rc != 0:
        log.error(error_message)
        return rc

    operator_repository_url = (
        f"https://{git_user}:{github_token}@github.com"
        + f"/{operator_repository}.git"
    )

    rc, operator_directory, error_message = clone_repository(
        operator_repository_url, operator_git_tag, base_directory, debug
    )
    if rc != 0:
        log.error(error_message)
        return rc

    rc, stdout, stderr = run(
        f"cd {operators_directory} && git checkout -b {operators_branch}",
        debug=debug,
    )
    if rc != 0:
        return (
            rc,
            f"Failed to create kudobuilder/operators branch: "
            + f"{operators_branch}\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    rc, stdout, stderr = run(
        f"cd {operator_directory} && git show -s --format='%an' {operator_git_tag}",
        debug=debug,
    )
    if rc != 0:
        return (
            rc,
            f"Failed to get git user name from "
            + f"{operator_repository}@{operator_git_tag}"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )
    git_user_name = stdout.strip()

    rc, stdout, stderr = run(
        f"cd {operator_directory} && git show -s --format='%ae' {operator_git_tag}",
        debug=debug,
    )
    if rc != 0:
        return (
            rc,
            f"Failed to get git user email from "
            + f"{operator_repository}@{operator_git_tag}"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )
    git_user_email = stdout.strip()

    rc, error_message = configure_git_user(
        operators_directory, git_user_name, git_user_email, debug
    )
    if rc != 0:
        log.error(error_message)
        return rc

    versioned_operator_directory = (
        f"{operators_directory}/repository/{operator_name}/{operator_git_tag}"
    )

    command = " && ".join(
        [
            f"mkdir -p {versioned_operator_directory}",
            f"cp -r {operator_directory}/operator {versioned_operator_directory}",
            f"cp -r {operator_directory}/docs {versioned_operator_directory}",
        ]
    )

    rc, stdout, stderr = run(command, debug=debug)
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"cd {operators_directory} && git add .", debug=debug
    )
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"cd {operators_directory} && git commit -am '{git_commit_message}'",
        debug=debug,
    )
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"cd {operators_directory} && git log -n 1", debug=debug
    )
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"cd {operators_directory} && git push origin {operators_branch}",
        debug=debug,
    )
    if rc != 0:
        return (
            rc,
            f"Failed to push '{operators_branch}' to '{operators_repository}'"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    success, http_response = create_pull_request(
        operators_repository,
        operators_base_branch,
        operators_branch,
        git_commit_message,
        "",
        github_token,
        PROGRAM_NAME,
        debug,
    )
    if not success:
        log.error(
            f"Error creating pull request\n"
            + f"response body: {http_response.read()}\n"
            + f"response headers: {http_response.getheaders()}"
        )
        return 1

    pull_request_url = json.loads(http_response.read())["html_url"]

    log.info(f"Successfully created PR: '{pull_request_url}'")

    return 0


if __name__ == "__main__":
    sys.exit(main())
