#!/usr/bin/env python3

from http import HTTPStatus
from pathlib import Path
from typing import Tuple, Optional, List
from urllib.parse import urlparse
import http
import http.client
import json
import logging
import os.path as path
import subprocess
import uuid


log = logging.getLogger(__name__)


def random_short_string() -> str:
    return uuid.uuid4().hex[:8]


def run(
    command: str,
    debug: bool = False,
    check: bool = False,
    timeout_seconds: Optional[int] = None,
) -> Tuple[int, str, str]:
    result = subprocess.run(
        [command],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        shell=True,
        check=check,
        timeout=timeout_seconds,
    )

    if result.stdout:
        stdout = result.stdout.decode("utf-8")
    else:
        stdout = ""

    if result.stderr:
        stderr = result.stderr.decode("utf-8")
    else:
        stderr = ""

    if debug:
        log.info(
            "Command '{}' exited with '{}'".format(command, result.returncode)
        )
        if stdout:
            log.info("stdout:\n{}".format(stdout))
        if stderr:
            log.info("stderr:\n{}".format(stderr))

    return result.returncode, stdout, stderr


def repository_dirty(debug: bool) -> bool:
    rc, stdout, stderr = run("git diff --quiet", debug=debug)
    return rc != 0


def remote_exists(remote: str, debug: bool) -> bool:
    rc, stdout, stderr = run(f"git remote show {remote}", debug=debug)
    return rc == 0


def matching_remote_branches(
    remote: str, branch: str, debug: bool
) -> (int, List[str], str):
    rc, stdout, stderr = run("git branch -r", debug=debug)
    if rc != 0:
        return (
            rc,
            [],
            "Error listing remote branches:"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    remote_branches = [l.strip() for l in stdout.split("\n") if l]

    return (
        0,
        [
            b
            for b in remote_branches
            if branch_exists_in_remote(b, remote, branch)
        ],
        "",
    )


def local_tag_exists(tag: str, debug: bool) -> bool:
    rc, stdout, stderr = run(f"git rev-parse refs/tags/{tag}", debug=debug)
    return rc == 0


def remote_tag_exists(remote: str, tag: str, debug: bool) -> bool:
    rc, stdout, stderr = run(
        f"git ls-remote --tags {remote} refs/tags/{tag}", debug=debug
    )
    return rc == 0 and "refs/tags/{tag}" in stdout


def branch_exists_in_remote(
    remote_branch: str, remote: str, branch: str
) -> bool:
    return remote_branch == f"{remote}/{branch}"


def local_branch_matches_remote_branch(
    remote: str, branch: str, debug: bool
) -> bool:
    rc, stdout, stderr = run(
        f"git diff {branch}...{remote}/{branch} --quiet", debug=debug
    )
    return rc == 0


def create_local_tag(tag: str, debug: str) -> (int, str):
    rc, stdout, stderr = run(f"git tag {tag}", debug=debug)
    if rc != 0:
        return (
            rc,
            f"Error creating local tag '{tag}'"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    return 0, ""


def push_tag(remote: str, tag: str, debug: str) -> (int, str):
    rc, stdout, stderr = run(f"git push {remote} {tag}", debug=debug)
    if rc != 0:
        return (
            rc,
            f"Error pushing tag '{tag}'\nstdout:\n{stdout}\nstderr:\n{stderr}",
        )

    return 0, ""


def configure_git_user(
    repository_directory: str, user_name: str, user_email: str, debug: str
) -> (int, str):
    command = " && ".join(
        [
            f"cd {repository_directory}",
            f"git config --local user.name '{user_name}'",
            f"git config --local user.email '{user_email}'",
        ]
    )

    rc, stdout, stderr = run(command, debug=debug)
    if rc != 0:
        error_message = (
            f"Error configuring git user '{user_name} <{user_email}>' "
            + f"for repository in '{repository_directory}'"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}"
        )

        return rc, error_message

    return 0, ""


def clone_repository(
    repository_url: str, reference: str, base_directory: str, debug: bool
) -> (int, str, str):
    repository = Path(urlparse(repository_url).path).stem
    target_directory = path.join(base_directory, repository.split("/")[-1])

    rc, stdout, stderr = run(
        f"git clone --depth 1 --branch {reference} "
        + f"{repository_url} {target_directory}",
        debug=debug,
    )
    if rc != 0:
        error_message = (
            f"Error cloning '{repository_url}@{reference}':"
            + f"\nstdout:\n{stdout}\nstderr:\n{stderr}"
        )

        return rc, target_directory, error_message

    return 0, target_directory, ""


def create_pull_request(
    repository: str,
    base_branch: str,
    branch: str,
    title: str,
    description: str,
    github_token: str,
    user_agent: str,
    debug: bool,
) -> (bool, http.client.HTTPResponse):
    headers = {
        "User-Agent": user_agent,
        "Content-Type": "application/json",
        "Authorization": f"Token {github_token}",
    }
    payload = {
        "title": title,
        "head": branch,
        "base": base_branch,
        "body": description,
    }
    connection = http.client.HTTPSConnection("api.github.com")
    url = "/repos/{}/pulls".format(repository)
    method = "POST"
    body = json.dumps(payload).encode("utf-8")

    if debug:
        log.info(f"HTTP request: {method} {url}\n{headers}\n\n{body}")

    connection.request(method, url, body=body, headers=headers)
    response = connection.getresponse()

    if debug:
        log.info(f"HTTP response: {response.status}")

    success = response.status in [HTTPStatus.CREATED]
    return success, response