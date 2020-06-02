# KUDO Cassandra Operator Development

**Table of Contents**

- [Requirements](#requirements)
- [Walkthrough](#walkthrough)
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
  - [Development cycle](#development-cycle)
  - [Release workflow](#release-workflow)
  - [Backport workflow](#backport-workflow)
  - [Additional release work](#additional-release-work)
  - [Snapshots (to be implemented)](#snapshots-to-be-implemented)
- [Synchronize changes to [kudobuilder/operators](https://github.com/kudobuilder/operators)](#synchronize-changes-to-kudobuilderoperatorshttpsgithubcomkudobuilderoperators)

## Requirements

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

## Walkthrough

### Cloning the git repository

Remove the `--recurse-submodules` flag if you don't have access to the private
[CI repository](https://github.com/mesosphere/data-services-ci). If you do have
access to it, also make sure you have
[GitHub SSH access](https://help.github.com/en/articles/connecting-to-github-with-ssh)
configured.

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
- (Only for macOS) [coreutils](https://formulae.brew.sh/formula/coreutils) with
  normal names
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
- [Prettier](https://prettier.io/)
- [pytablewriter](https://pytablewriter.readthedocs.io/en/latest/)
- [Python 3](https://docs.python.org/3/)
- [shellcheck](https://www.shellcheck.net/)
- [docker](https://docs.docker.com/)

### Compiling templates

The KUDO Cassandra operator makes use of templates so that data that needs to be
present in files or available as environment variables in scripts can be
centralized in a single place and templated into files or loaded into
environment variables in scripts.

The centralized place for that data is the `metadata.sh` file.

Templates are under the `templates` directory. The `tools/compile_templates.sh`
script will compile all templates under `templates` to files in the repository.

**note:** these scripts currently only run on linux. On other platforms pass the
the script through the `tools/docker.sh` script, for example:
`./tools/docker.sh ./tools/compile_templates.sh`.

For example, given the following data file:

`metadata.sh`

```bash
export CASSANDRA_VERSION="3.11.5"
```

Running `./tools/compile_templates.sh` will compile

`templates/operator/operator.yaml.template`

```yaml
apiVersion: kudo.dev/v1beta1
name: "cassandra"
appVersion: "${CASSANDRA_VERSION}"
```

into

`operator/operator.yaml`

```yaml
apiVersion: kudo.dev/v1beta1
name: "cassandra"
appVersion: "3.11.5"
```

Assuming also there's a script named `cassandra_version.sh` that looks like

```bash
source "../metadata.sh"
echo "${CASSANDRA_VERSION}"
```

Running it will output `3.11.5`.

```bash
$ ./cassandra_version.sh
3.11.5
```

It's important to notice that **changes to templated files have to be done on
the template and compiled**. For example, changes to `operator/operator.yaml`
have to made in `templates/operator/operator.yaml.template` and then compiled
with `tools/compile_templates.sh`.

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
[Semantic Versioning 2.0.0](https://semver.org/). The _version_ is composed of
the underlying Apache Cassandra version (_app version_) concatenated with the
_operator version_. For example, in `3.11.4-0.1.0`, `3.11.4` is the Apache
Cassandra version and `0.1.0` is the operator version. **The operator version is
reset on every minor app version release.** The current reset target is `0.1.0`.
Eventually, the reset target will become `1.0.0`.

##### Example hypothetical timeline for releases

| Time | Apache C\* version | Operator version | KUDO API version | Comment                                    | Change                      |
| ---- | ------------------ | ---------------- | ---------------- | ------------------------------------------ | --------------------------- |
| T0   | 3.11.4             | 0.1.0            | v1beta1          | Initial release based on Apache C\* 3.11.x | -                           |
| T1   | 3.11.4             | 0.1.1            | v1beta1          | Bug fix in operator-related code           | Operator patch version bump |
| T2   | 3.12.0             | 0.1.0            | v1beta1          | Apache C\* 3.12.x release                  | Operator version reset      |
| T3   | 3.11.4             | 0.2.0            | v1beta1          | Operator-related feature A added to 3.11.x | Operator minor version bump |
| T3   | 3.12.0             | 0.3.0            | v1beta1          | Operator-related feature A added to 3.12.x | Operator minor version bump |
| T4   | 4.0.0              | 0.1.0            | v1beta1          | Apache C\* 4.0.x release                   | Operator version reset      |
| T5   | 3.11.4             | 0.3.0            | v1beta1          | Operator-related feature B added to 3.11.x | Operator minor version bump |
| T5   | 3.12.0             | 0.4.0            | v1beta1          | Operator-related feature B added to 3.12.x | Operator minor version bump |
| T5   | 4.0.0              | 0.2.0            | v1beta1          | Operator-related feature B added to 4.0.x  | Operator minor version bump |
| T6   | 3.11.4             | 1.0.0            | v1               | KUDO API version change                    | Operator major version bump |
| T6   | 3.12.0             | 1.0.0            | v1               | KUDO API version change                    | Operator major version bump |
| T6   | 4.0.0              | 1.0.0            | v1               | KUDO API version change                    | Operator major version bump |

It's important to note that **operator versions for different app versions are
unrelated**. e.g., in the example above both `3.11.4-0.1.0` and `4.0.0-0.1.0`
have `0.1.0` as the _operator version_, but wouldn't necessarily share any
commonality with regards to the operator itself. The operator version
progression is only meaningful within an app version's `major.minor` family,
i.e. `3.11.x` and `4.0.x`.

### Release workflow

#### Feature development

Development happens in feature branches which are merged into the `master`
branch via GitHub PRs.

#### Stable branch creation

When it is decided that a release needs to be done, a _stable branch_ is created
based off of the `master` branch. This can simply be done in GitHub web UI using
the branch selector widget:

![](docs/images/branch.png =300x)

The name of stable branch is typically `release-vx.y` where `x.y` is the
Cassandra `major.minor` version.

In this branch all operator dependencies (Docker images, KUDO version, Golang
libraries, etc.) are made to be _stable_, as in no _running versions_ (SNAPSHOT,
latest, etc.) are used.

This is achieved by creating and merging a PR _against the stable branch_ where:

1.  the value of `POSSIBLE_SNAPHOT_SUFFIX` in [`metadata.sh`](metadata.sh) is
    set to an empty string,
1.  (as needed) the various `*_VERSION` variables are set as necessary for base
    tech and the operator version, according to the versioning scheme shown
    above,
1.  necessary files are updated by running `./tools/compile_template.sh`

#### Creating Release Notes

Create an entry for the release in the [CHANGELOG](CHANGELOG.md).

#### Tagging release

A tag is created for the release, using
[the GitHub UI](https://github.com/mesosphere/kudo-cassandra-operator/releases/new).
Copy the contents of the release entry from the changelog updated above.

![](docs/images/tag.png =300x)

#### Building docker images

Docker images need to be built and pushed with names matching the tagged
release. This is typically achieved by running the
[`images/build.sh`](./images/build.sh) script using TeamCity on the release tag
with an additional parameter:

1. Click the `...` next to the `Run` button on the
   [Docker Push](https://teamcity.mesosphere.io/buildConfiguration/Frameworks_DataServices_Kudo_Cassandra_Tools_DockerPush)
   build configuration page.
1. Select the release tag on the `Changes` tab: ![](docs/images/run-on-tag.png =300x)
1. Add an `env.DISABLE_IMAGE_DISAMBIGUATION_SUFFIX` environment variable on the
   `Parameters` tab: ![](docs/images/run-with-param.png =300x)
1. click `Run Build`

#### Copying to the `kudobuilder/operators` repository

The [kudobuilder/operators](https://github.com/kudobuilder/operators) repository
contains a collection of KUDO operators. As of right now (2019-12-11) it is
required that operators are published there so that packages can be built for
installation via `kubectl kudo install`.

The
[`tools/create_operators_pull_request.py`](tools/create_operators_pull_request.py)
script copies over all "KUDO operator"-related files to kudobuilder.operators.

For example, the following command creates a PR under kudobuilder/operators
copying all "KUDO operator"-related files from the
mesosphere/kudo-cassandra-operator (the [`operator`](operator) and
[`docs`](docs) directory as of 2019-12-11) into a directory under
kudobuilder/operators:

```bash
./tools/create_operators_pull_request.py \
  --operator-repository mesosphere/kudo-cassandra-operator \
  --operator-name cassandra \
  --operator-git-tag v3.11.5-0.1.1 \
  --github-token "${github_token}"
```

#### Building and pushing a KUDO operator package

Ask the friendly folks on #kudo channel on the kubernetes.slack.com instance.

#### Lather, rinse, repeat as required

Once the stable branch is created, additional commits may be landed on it either
via merging from the `master` branch or cherry-picking individual commits.

These can then be released by updating the `OPERATOR_VERSION` in
[`metadata.sh`](metadata.sh) and repeating the last steps of this workflow.
