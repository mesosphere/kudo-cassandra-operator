#!/usr/bin/env python3

from os import path
import argparse
import logging
import re
import sys

from utils import (
    run,
    repository_dirty,
    remote_exists,
    matching_remote_branches,
    local_tag_exists,
    remote_tag_exists,
    local_branch_matches_remote_branch,
    create_local_tag,
    push_tag,
)

__directory__ = path.dirname(__file__)

log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%d %H:%M:%SZ",
)

SEMVER_MAJOR_MINOR_PATTERN = "(0|[1-9]\d*)\.(0|[1-9]\d*)"

# https://regex101.com/r/vkijKf/1/
SEMVER_PATTERN = "(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?"

RELEASE_TAG_PATTERN = f"v{SEMVER_PATTERN}-{SEMVER_PATTERN}"

STABLE_BRANCH_NAME_PATTERN = f"release-v{SEMVER_MAJOR_MINOR_PATTERN}"


def valid_git_tag(tag: str) -> bool:
    return bool(re.match(RELEASE_TAG_PATTERN, tag))


def valid_stable_branch_name(branch: str) -> bool:
    return bool(re.match(STABLE_BRANCH_NAME_PATTERN, branch))


def validate_arguments_and_environment(
    git_remote: str, git_branch: str, git_tag: str, debug: bool
) -> int:
    if repository_dirty(debug):
        log.info(
            "Local repository is dirty. "
            + "Make sure it is clean and run the script again"
        )
        return 1

    if not remote_exists(git_remote, debug):
        log.info(
            f"Invalid stable branch name: '{git_branch}'. Stable branch names "
            + f"should follow the pattern: {STABLE_BRANCH_NAME_PATTERN}"
        )
        return 1

    if not valid_stable_branch_name(git_branch):
        log.info(
            f"Invalid stable branch name: '{git_branch}'. Stable branch names "
            + f"should follow the pattern: {STABLE_BRANCH_NAME_PATTERN}"
        )
        return 1

    if not valid_git_tag(git_tag):
        log.info(
            f"Invalid release tag: {git_tag}. Release tags should follow the "
            + f"pattern: {RELEASE_TAG_PATTERN}"
        )
        return 1

    rc, stdout, stderr = run(
        f"git fetch --prune --tags {git_remote}", debug=debug
    )
    if rc != 0:
        log.error(
            "Error fetching remote '{git_remote}':"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}"
        )
        return rc

    rc, _matching_remote_branches, error_message = matching_remote_branches(
        git_remote, git_branch, debug
    )
    if rc != 0:
        log.error(error_message)
        return rc

    if len(_matching_remote_branches) == 0:
        log.warning(
            f"Didn't find remote branch '{git_remote}/{git_branch}'. "
            + "Push a stable branch so that a release can be created from it"
        )
        return 1
    elif len(_matching_remote_branches) == 1:
        log.info(f"Found remote branch '{_matching_remote_branches[0]}'")
    else:
        log.warning(
            "Found more than one remote branch matching "
            + f"'{git_remote}/{git_branch}'"
        )
        return 1

    if local_tag_exists(git_tag, debug):
        log.error(
            f"Local tag already exists: '{git_tag}'. "
            + "Can't release a version when a tag "
            + "has potentially already been published"
        )
        return 1

    if remote_tag_exists(git_remote, git_tag, debug):
        log.error(
            f"Remote tag already exists: 'refs/tags/{git_tag}'. "
            + "Can't release a version when a remote tag "
            + "already has been published"
        )
        return 1

    rc, stdout, stderr = run(
        f"git checkout -b {git_branch} {git_remote}/{git_branch}", debug=debug
    )
    if rc != 0:
        if "already exists" in stderr:
            if not local_branch_matches_remote_branch(
                git_remote, git_branch, debug
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
    parser = argparse.ArgumentParser(
        description="Releases the KUDO Cassandra Operator"
    )

    parser.add_argument("--git-remote", type=str, default="origin", help="")
    parser.add_argument("--git-branch", type=str, help="")
    parser.add_argument("--git-tag", type=str, help="")
    parser.add_argument("--debug", action="store_true", default=False, help="")

    args = parser.parse_args()
    git_remote = args.git_remote
    git_branch = args.git_branch
    git_tag = args.git_tag
    debug = args.debug

    rc = validate_arguments_and_environment(
        git_remote, git_branch, git_tag, debug
    )
    if rc != 0:
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
    sys.exit(main())
