# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Apache 2.0 license
  ([c792f72d](https://github.com/mesosphere/kudo-cassandra-operator/commit/c792f72d132ad01dd02859f3dc266f3e54142e32))

## [3.11.4-0.1.0] - 2019-11-13

### Added

- Configurable `cassandra.yaml` and `jvm.options` parameters
  ([#1](https://github.com/mesosphere/kudo-cassandra-operator/pull/1),
  [#9](https://github.com/mesosphere/kudo-cassandra-operator/pull/9))
- JVM memory locking out of the box
- Prometheus metrics and Grafana dashboard
  ([#4](https://github.com/mesosphere/kudo-cassandra-operator/pull/4))
- Horizontal scaling
- Rolling parameter updates
- Readiness probe
  ([#6](https://github.com/mesosphere/kudo-cassandra-operator/pull/6))
- Unpriviledged container execution

[unreleased]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.4-0.1.0...HEAD
[3.11.4-0.1.0]:
  https://github.com/mesosphere/kudo-cassandra-operator/releases/tag/v3.11.4-0.1.0
