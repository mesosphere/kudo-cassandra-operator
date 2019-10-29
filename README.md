# KUDO Cassandra Operator

## Requirements

- [GitHub SSH
  access](https://help.github.com/en/articles/connecting-to-github-with-ssh)
- Docker ([macOS](https://docs.docker.com/docker-for-mac/),
  [Ubuntu](https://docs.docker.com/install/linux/docker-ce/ubuntu/). Last tested
  on 19.03.2)
- [Docker daemon running under a non-root
  user](https://docs.docker.com/install/linux/linux-postinstall/) (only for
  Linux)
- [KUDO](https://github.com/kudobuilder/kudo/releases) (check `KUDO_VERSION` in
  `metadata.sh` to see the last tested version)
- Kubernetes cluster (last tested on
  [Konvoy](https://github.com/mesosphere/konvoy/releases)) ((check
  `KUBERNETES_VERSION` in `metadata.sh` to see the last tested version))

## Installing

### Cloning the git repository

```bash
git clone --recurse-submodules git@github.com:mesosphere/kudo-cassandra-operator.git /path/to/kudo-cassandra-operator
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
