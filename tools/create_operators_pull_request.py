#!/usr/bin/env python3

from datetime import datetime
from typing import Tuple
import argparse
import json
import logging
import os.path
import sys
import tempfile

from utils import (
    run,
    get_git_version,
    get_git_user,
    get_sha,
    github_repository_url,
    authenticated_github_repository_url,
    clone_repository,
    random_short_string,
    configure_git_user,
    create_pull_request,
)
from release import RELEASE_TAG_PATTERN

PROGRAM_NAME = os.path.basename(__file__)

log = logging.getLogger(__name__)


OPERATORS_REPOSITORY = "kudobuilder/operators"


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Open a PR copying a KUDO Operator's files to the "
        + f"{OPERATORS_REPOSITORY} repository"
    )

    parser.add_argument(
        "--operator-repository",
        type=str,
        required=True,
        help="The KUDO Operator repository to copy the files from "
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
        help="The git tag in the KUDO operator repository to copy the files "
        + f"from. This is also the directory name in {OPERATORS_REPOSITORY}, "
        + "e.g., repository/cassandra/$OPERATOR_GIT_TAG",
    )
    parser.add_argument(
        "--github-token",
        type=str,
        required=True,
        help="The GitHub token used for opening the PR against "
        + f"{OPERATORS_REPOSITORY}",
    )
    parser.add_argument(
        "--git-commit-message",
        type=str,
        help=f"The git commit message in the {OPERATORS_REPOSITORY} PR branch. "
        + "This will also be the PR title",
    )
    parser.add_argument(
        "--operators-base-branch",
        type=str,
        default="master",
        help=f"The {OPERATORS_REPOSITORY} branch to open a PR against",
    )
    parser.add_argument(
        "--operators-repository",
        type=str,
        default=OPERATORS_REPOSITORY,
        help=f"Use a different repository than {OPERATORS_REPOSITORY}",
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

    return parser.parse_args()


def prepare_git_repositories(
    base_directory: str,
    operators_repository: str,
    operators_base_branch: str,
    operators_branch: str,
    operator_repository: str,
    operator_git_ref: str,
    git_user: str,
    github_token: str,
    debug: bool,
) -> Tuple[int, str, str, str]:
    """
    1. Clones the "operators collection" repository (e.g., kudobuilder/operators)
    2. Clones the operator repository (e.g., mesosphere/kudo-cassandra-operator)
    3. Creates desired branch in the "operators collection" repository
    4. Configures the git user name and email in the "operators collection" repository to be the git user name and email at the operator repository's `operator_git_ref`.

    Returns the  "operators collection" and operator directories.
    """

    operators_repository_url = authenticated_github_repository_url(
        git_user, github_token, operators_repository
    )

    operator_repository_url = authenticated_github_repository_url(
        git_user, github_token, operator_repository
    )

    rc, operators_directory, error_message = clone_repository(
        operators_repository_url, operators_base_branch, base_directory, debug
    )
    if rc != 0:
        return rc, error_message, "", ""

    rc, operator_directory, error_message = clone_repository(
        operator_repository_url, operator_git_ref, base_directory, debug
    )
    if rc != 0:
        return rc, error_message, "", ""

    rc, stdout, stderr = run(
        f"git -C {operators_directory} checkout -b {operators_branch}",
        debug=debug,
    )
    if rc != 0:
        return (
            rc,
            f"Failed to create {OPERATORS_REPOSITORY} branch: "
            + f"{operators_branch}\nstdout:\n{stdout}\nstderr:\n{stderr}",
            "",
            "",
        )

    rc, error_message, git_user_name, git_user_email = get_git_user(
        operator_directory, operator_git_ref, debug
    )
    if rc != 0:
        return rc, error_message, "", ""

    rc, error_message = configure_git_user(
        operators_directory, git_user_name, git_user_email, debug
    )
    if rc != 0:
        return rc, error_message, "", ""

    return 0, "", operators_directory, operator_directory


def build_versioned_operator_directory(
    operators_directory: str, operator_name: str, operator_git_tag: str
) -> str:
    """Returns a versioned operator directory for the "operators collection"
    repository.

    e.g., /kudobuilder-operators/repository/cassandra/3.11
    """
    match = RELEASE_TAG_PATTERN.match(operator_git_tag)
    # Assuming APP_VERSION follows SemVer for now.
    app_version_major_minor = f"{match[1]}.{match[2]}"

    return f"{operators_directory}/repository/{operator_name}/{app_version_major_minor}"


def commit_copied_operator_files_and_push_branch(
    operators_directory: str,
    operators_repository: str,
    operators_branch: str,
    operator_directory: str,
    operator_name: str,
    operator_git_tag: str,
    git_commit_message: str,
    debug: bool,
) -> Tuple[int, str]:
    """Copies, commits and pushes operator-related files from the operator
    repository (e.g., mesosphere/kudo-cassandra-operator) directory into the
    "operators collection" repository (e.g., kudobuilder/operators)
    directory."""

    versioned_operator_directory = build_versioned_operator_directory(
        operators_directory, operator_name, operator_git_tag
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

    rc, stdout, stderr = run(f"git -C {operators_directory} add .", debug=debug)
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"git -C {operators_directory} commit -am '{git_commit_message}'",
        debug=debug,
    )
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"git -C {operators_directory} diff HEAD^...HEAD", debug=debug
    )
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"git -C {operators_directory} log -n 1", debug=debug
    )
    if rc != 0:
        return rc, f"stdout:\n{stdout}\nstderr:\n{stderr}"

    rc, stdout, stderr = run(
        f"git -C {operators_directory} push origin {operators_branch}",
        debug=debug,
    )
    if rc != 0:
        return (
            rc,
            f"Failed to push '{operators_branch}' to '{operators_repository}'"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    return 0, ""


def automated_operators_repository_commit_message(
    operator_repository: str,
    operator_name: str,
    operator_directory: str,
    operator_git_tag: str,
    debug: bool,
) -> Tuple[int, str, str, str]:
    rc, stdout, stderr = get_sha(operator_directory, debug)
    if rc != 0:
        return (
            rc,
            "",
            "",
            f"Failed to get git SHA from '{operator_directory}'"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    operator_git_sha = stdout.strip()
    operator_repository_url = github_repository_url(operator_repository)
    operator_git_tag_url = (
        f"{operator_repository_url}/releases/tag/{operator_git_tag}"
    )
    operator_git_sha_url = (
        f"{operator_repository_url}/commit/{operator_git_sha}"
    )

    commit_message_subject = (
        f"Release {operator_name} {operator_git_tag} (automated commit)."
    )
    commit_message_body = "\n".join(
        [
            f"| | |",
            f"|-|-|",
            f"| Repository | {operator_repository_url} |",
            f"| Operator | {operator_name} |",
            f"| Git tag | {operator_git_tag_url} |",
            f"| Git SHA | {operator_git_sha_url} |",
            f"| Date (UTC) | {datetime.utcnow()} |",
        ]
    )

    return 0, commit_message_subject, commit_message_body, ""


def automated_operators_repository_branch(
    operator_name: str, operator_git_tag: str
) -> str:
    return f"{operator_name}_{operator_git_tag}_{random_short_string()}"


def main() -> int:
    args = parse_arguments()

    operator_repository = args.operator_repository
    operator_name = args.operator_name
    operator_git_tag = args.operator_git_tag
    github_token = args.github_token
    operators_base_branch = args.operators_base_branch
    operators_repository = args.operators_repository
    git_user = args.git_user
    debug = args.debug

    operators_branch = automated_operators_repository_branch(
        operator_name, operator_git_tag
    )

    rc, stdout, stderr = get_git_version(debug)
    if rc != 0:
        log.error(error_message)
        return rc
    git_version = stdout.strip()

    log.info(git_version)

    with tempfile.TemporaryDirectory(prefix="kudo_dev_") as base_directory:
        (
            rc,
            error_message,
            operators_directory,
            operator_directory,
        ) = prepare_git_repositories(
            base_directory,
            operators_repository,
            operators_base_branch,
            operators_branch,
            operator_repository,
            operator_git_tag,
            git_user,
            github_token,
            debug,
        )
        if rc != 0:
            log.error(error_message)
            return rc

        if args.git_commit_message:
            git_commit_message_subject = args.git_commit_message
            git_commit_message_body = ""
        else:
            (
                rc,
                git_commit_message_subject,
                git_commit_message_body,
                error_message,
            ) = automated_operators_repository_commit_message(
                operator_repository,
                operator_name,
                operator_directory,
                operator_git_tag,
                debug,
            )
            if rc != 0:
                log.error(error_message)
                return rc

        git_commit_message = (
            f"{git_commit_message_subject}\n\n{git_commit_message_body}"
        ).strip()

        rc, error_message = commit_copied_operator_files_and_push_branch(
            operators_directory,
            operators_repository,
            operators_branch,
            operator_directory,
            operator_name,
            operator_git_tag,
            git_commit_message,
            debug,
        )
        if rc != 0:
            log.error(error_message)
            return rc

        success, http_response = create_pull_request(
            operators_repository,
            operators_base_branch,
            operators_branch,
            git_commit_message_subject,
            git_commit_message_body,
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
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
        datefmt="%Y-%m-%d %H:%M:%SZ",
    )
    sys.exit(main())
