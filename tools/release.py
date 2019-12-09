#!/usr/bin/env python3

import argparse
import logging
import re
import sys
import tempfile

from utils import (
    run,
    authenticated_github_repository_url,
    repository_dirty,
    remote_exists,
    get_matching_remote_branches,
    local_tag_exists,
    remote_tag_exists,
    local_branch_matches_remote_branch,
    get_git_user,
    configure_git_user,
    clone_repository,
    create_local_tag,
    push_tag,
)

log = logging.getLogger(__name__)

SEMVER_MAJOR_MINOR_PATTERN = re.compile(r"(0|[1-9]\d*)\.(0|[1-9]\d*)")

# https://regex101.com/r/vkijKf/1/
SEMVER_PATTERN = re.compile(
    r"(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?"
)

RELEASE_TAG_PATTERN = re.compile(
    f"v{SEMVER_PATTERN.pattern}-{SEMVER_PATTERN.pattern}"
)

STABLE_BRANCH_NAME_PATTERN = re.compile(
    f"release-v{SEMVER_MAJOR_MINOR_PATTERN.pattern}"
)


def valid_release_tag(tag: str) -> bool:
    return bool(RELEASE_TAG_PATTERN.match(tag))


def valid_stable_branch_name(branch: str) -> bool:
    return bool(STABLE_BRANCH_NAME_PATTERN.match(branch))


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Release a KUDO Operator")

    parser.add_argument(
        "--repository",
        type=str,
        required=True,
        help="The KUDO Operator repository to be released "
        + "(e.g., mesosphere/kudo-cassandra-operator)",
    )
    parser.add_argument(
        "--git-branch",
        type=str,
        required=True,
        help="The name of the KUDO Operator repository git branch",
    )
    parser.add_argument(
        "--git-tag",
        type=str,
        required=True,
        help="The desired git tag that will be created in the KUDO Operator "
        + "repository from the head of the GIT_BRANCH",
    )
    parser.add_argument(
        "--git-remote",
        type=str,
        default="origin",
        help="The name of the git remote for the KUDO Operator repository",
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


def validate_arguments_and_environment(
    repository_directory: str,
    git_remote: str,
    git_branch: str,
    git_tag: str,
    debug: bool,
) -> int:
    if not valid_stable_branch_name(git_branch):
        log.error(
            f"Invalid stable branch name: '{git_branch}'. Stable branch names "
            + f"should follow the pattern: {STABLE_BRANCH_NAME_PATTERN.pattern}"
        )
        return 1

    if not valid_release_tag(git_tag):
        log.error(
            f"Invalid release tag: '{git_tag}'. Release tags should follow the "
            + f"pattern: {RELEASE_TAG_PATTERN.pattern}"
        )
        return 1

    if repository_dirty(repository_directory, debug):
        log.error(
            "Local repository is dirty. "
            + "Make sure it is clean and run the script again"
        )
        return 1

    if not remote_exists(repository_directory, git_remote, debug):
        log.error(f"Remote '{git_remote}' doesn't exist")
        return 1

    rc, stdout, stderr = run(
        f"git -C {repository_directory} fetch --prune --tags {git_remote}",
        debug=debug,
    )
    if rc != 0:
        log.error(
            "Error fetching remote '{git_remote}':"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}"
        )
        return rc

    rc, matching_remote_branches, error_message = get_matching_remote_branches(
        repository_directory, git_remote, git_branch, debug
    )
    if rc != 0:
        log.error(error_message)
        return rc

    if len(matching_remote_branches) == 0:
        log.warning(
            f"Didn't find remote branch '{git_remote}/{git_branch}'. "
            + "Push a stable branch so that a release can be created from it"
        )
        return 1
    elif len(matching_remote_branches) == 1:
        log.info(f"Found remote branch '{matching_remote_branches[0]}'")
    else:
        log.warning(
            "Found more than one remote branch matching "
            + f"'{git_remote}/{git_branch}'"
        )
        return 1

    if local_tag_exists(repository_directory, git_tag, debug):
        log.error(
            f"Local tag already exists: '{git_tag}'. "
            + "Can't release a version when a tag "
            + "has potentially already been published"
        )
        return 1

    if remote_tag_exists(repository_directory, git_remote, git_tag, debug):
        log.error(
            f"Remote tag already exists: 'refs/tags/{git_tag}'. "
            + "Can't release a version when a remote tag has "
            + "already been published"
        )
        return 1

    rc, stdout, stderr = run(
        f"git -C {repository_directory} "
        + f"checkout -b {git_branch} {git_remote}/{git_branch}",
        debug=debug,
    )
    if rc != 0:
        if "already exists" in stderr:
            if not local_branch_matches_remote_branch(
                repository_directory, git_remote, git_branch, debug
            ):
                log.error(
                    f"Local branch '{git_branch}' doesn't match "
                    + f"remote branch '{git_remote}/{git_branch}'"
                )
                return rc
        else:
            log.error(
                f"Error checking out local '{git_remote}' branch from "
                + f"'{git_remote}/{git_branch}'"
                + f"\nstdout:\n{stdout}\nstderr:\n{stderr}"
            )
            return rc

    return 0


def main() -> int:
    args = parse_arguments()

    repository = args.repository
    git_branch = args.git_branch
    git_tag = args.git_tag
    git_remote = args.git_remote
    github_token = args.github_token
    git_user = args.git_user
    debug = args.debug

    repository_url = authenticated_github_repository_url(
        git_user, github_token, repository
    )

    with tempfile.mkdtemp("_kudo_dev") as base_directory:
        rc, directory, error_message = clone_repository(
            repository_url, git_tag, base_directory, debug
        )
        if rc != 0:
            log.error(error_message)
            return rc

        rc = validate_arguments_and_environment(
            directory, git_remote, git_branch, git_tag, debug
        )
        if rc != 0:
            return rc

        rc, error_message, git_user_name, git_user_email = get_git_user(
            directory, git_branch, debug
        )
        if rc != 0:
            log.error(error_message)
            return rc

        return 0

        rc, error_message = configure_git_user(
            directory, git_user_name, git_user_email, debug
        )
        if rc != 0:
            log.error(error_message)
            return rc

        rc, error_message = create_local_tag(git_tag, debug)
        if rc != 0:
            log.error(error_message)
            return rc

        rc, error_message = push_tag(git_remote, git_tag, debug)
        if rc != 0:
            log.error(error_message)
            return rc

        log.info(f"'{git_tag}' released successfully from '{git_branch}'")

        return 0


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
        datefmt="%Y-%m-%d %H:%M:%SZ",
    )
    sys.exit(main())
