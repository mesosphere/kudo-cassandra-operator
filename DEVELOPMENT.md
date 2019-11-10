# KUDO Cassandra Operator Development

**Table of Contents**

- [Requirements](#requirements)
- [Installing](#installing)
    - [Cloning the git repository](#cloning-the-git-repository)
    - [Choosing a name and a namespace for the instance](#choosing-a-name-and-a-namespace-for-the-instance)
    - [Creating the namespace (if it doesn't exist)](#creating-the-namespace-if-it-doesnt-exist)
    - [Installing the KUDO Cassandra operator](#installing-the-kudo-cassandra-operator)
        - [Checking out the deploy plan](#checking-out-the-deploy-plan)
        - [Checking out the pods](#checking-out-the-pods)
        - [Getting the 0th pod name](#getting-the-0th-pod-name)
        - [Checking out the output of `nodetool stats`](#checking-out-the-output-of-nodetool-stats)
        - [Running a `cassandra-stress` workload](#running-a-cassandra-stress-workload)
    - [Uninstalling the KUDO Cassandra operator](#uninstalling-the-kudo-cassandra-operator)
- [Development](#development)
    - [Additional requirements](#additional-requirements)
    - [Compiling templates](#compiling-templates)
    - [Running static code analyzers](#running-static-code-analyzers)
        - [Automatically formatting all files](#automatically-formatting-all-files)
        - [Checking if all files pass formatting and linting checks](#checking-if-all-files-pass-formatting-and-linting-checks)
    - [Building Docker images](#building-docker-images)
    - [Running tests](#running-tests)
    - [Style guide](#style-guide)
        - [Opening pull requests](#opening-pull-requests)
        - [Merging pull requests](#merging-pull-requests)
- [Releasing](#releasing)
    - [Versioning](#versioning)
    - [Development Cycle](#development-cycle)
    - [Release Workflow](#release-workflow)
    - [Backport Workflow](#backport-workflow)

## Requirements

- [GitHub SSH access](https://help.github.com/en/articles/connecting-to-github-with-ssh)
- Docker ([macOS](https://docs.docker.com/docker-for-mac/),
  [Ubuntu](https://docs.docker.com/install/linux/docker-ce/ubuntu/). Last tested
  on 19.03.2)
- [Docker daemon running under a non-root user](https://docs.docker.com/install/linux/linux-postinstall/)
  (only for Linux)
- [KUDO](https://github.com/kudobuilder/kudo/releases) (check `KUDO_VERSION` in
  `metadata.sh` to see the last tested version)
- Kubernetes cluster (last tested on
  [Konvoy](https://github.com/mesosphere/konvoy/releases)) (check
  `KUBERNETES_VERSION` in `metadata.sh` to see the last tested version)

## Installing

### Cloning the git repository

```bash
git clone \
    --recurse-submodules \
    git@github.com:mesosphere/kudo-cassandra-operator.git \
    /path/to/kudo-cassandra-operator
```

All commands assume that you're in the project root directory.

```bash
cd /path/to/kudo-cassandra-operator
```

### Choosing a name and a namespace for the instance

```bash
kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="default"
```

### Creating the namespace (if it doesn't exist)

```bash
kubectl create namespace "${kudo_cassandra_instance_namespace}"
```

### Installing the KUDO Cassandra operator

```bash
kubectl kudo install ./operator \
        --instance="${kudo_cassandra_instance_name}" \
        --namespace="${kudo_cassandra_instance_namespace}"
```

#### Checking out the deploy plan

```bash
kubectl kudo plan status \
        --instance="${kudo_cassandra_instance_name}" \
        --namespace="${kudo_cassandra_instance_namespace}"
```

#### Checking out the pods

```bash
kubectl get pods -n "${kudo_cassandra_instance_namespace}"
```

#### Getting the 0th pod name

```bash
kudo_cassandra_pod_0="$(kubectl get pods \
                                -o 'jsonpath={.items[0].metadata.name}' \
                                -n "${kudo_cassandra_instance_namespace}")"
```

#### Checking out the output of `nodetool stats`

```bash
kubectl exec "${kudo_cassandra_pod_0}" \
        -n "${kudo_cassandra_instance_namespace}" \
        -- \
        bash -c 'nodetool status'
```

#### Running a `cassandra-stress` workload

```bash
svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"
```

```bash
kubectl exec "${kudo_cassandra_pod_0}" \
        -n "${kudo_cassandra_instance_namespace}" \
        -- \
        bash -c "cassandra-stress write -node ${svc_endpoint}"
```

### Uninstalling the KUDO Cassandra operator

```bash
./scripts/uninstall_operator.sh \
  --instance "${kudo_cassandra_instance_name}" \
  --namespace "${kudo_cassandra_instance_namespace}"
```

## Development

### Additional requirements

- bash 4+ ([macOS](https://formulae.brew.sh/formula/bash))
- envsubst ([macOS](https://formulae.brew.sh/formula/gettext))
- [shellcheck](https://www.shellcheck.net/)
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)

### Compiling templates

```bash
./tools/compile_templates.sh
```

### Running static code analyzers

#### Automatically formatting all files

```bash
./tools/format_files.sh
```

#### Checking if all files pass formatting and linting checks

```bash
./tools/check_files.sh
```

### Building Docker images

```bash
./images/build.sh
```

### Running tests

### Style guide

#### Opening pull requests

PR titles should be in imperative mood, useful and concise. Example:

```
Add support for new thing.
```

PR descriptions should include additional context regarding what is achieved
with the PR, why is it needed, rationale regarding decisions that were made,
possibly with pointers to actual commits.

Example:

```
To make it possible for the new thing we had to:
- Prepare this other thing (5417f75)
- Clean up something else (ec4c78d)

This was required because of this and that.

Example output of thing:

    {
      "a": 2
    }


Please look into http://www.somewebsite.com/details-about-thing
for more context.
```

#### Merging pull requests

When all checks are green, a PR should be merged as a squash-commit, with its
message being the PR title followed by the PR number. Example:

```
Add support for new thing. (#42)
```

The description for the squash-commit will ideally be the PR description
verbatim. If the PR description was empty (it probably shouldn't have been!) the
squash-commit description will by default be a list of all the commits in the
PR's branch. That list should be cleaned up to only contain useful entries (no
`fix`, `formatting`, `changed foo`, `refactored bar`), or rewritten so that
additional context is added to the commit, like in the example above for PR
descriptions.

## Releasing

### Versioning

The current versioning scheme for the current KUDO Cassandra Operator follows
[Semantic Versioning 2.0.0](https://semver.org/). The _combined version_ is
composed of the underlying Apache Cassandra version (_app version_) concatenated
with the _operator version_. For example, in the combined version
`3.11.4-0.1.0`, `3.11.4` is the Apache Cassandra version and `0.1.0` is the
operator version.

### Development Cycle

Development happens in feature branches which are merged into the master branch
via GitHub PRs. When it is deciced that a release needs to be done, a _stable
branch_ is created based off of the master branch. In this branch all operator
dependencies (Docker images, KUDO version, Golang libraries, etc.) are made to
be _stable_, as in no SNAPSHOTs are used. The app version and operator version
are also updated to be the desired version to be released. After the git commits
doing the above are made, a git tag is created with the desired version to be
released.

### Release Workflow

A concrete example: it is desired that `3.11.4-0.1.0` is released:

1. A `release-v3.11` branch is created from master
1. Changes making dependencies stable and changing the app version be `3.11.4`
   and the operator version be `0.1.0` are committed and pushed to the remote
1. A `v3.11.4-0.1.0` git tag is created from the `release-v3.11` branch HEAD

The [release.py](./tools/release.py) script can be used to achieve the last step
above:

```bash
./tools/release.py --git-branch release-v3.11 --git-tag v3.11.4-0.1.0
```

### Backport Workflow

Any further releases based on Apache Cassandra `3.11.x` should originate from
backported changes from the master branch to the existing `release-v3.11`
branch, and then released to tags.
