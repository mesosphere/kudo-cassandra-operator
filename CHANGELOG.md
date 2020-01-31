# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

### Added

- Transport Encryption TLS for node-to-node and client-to-node connections.
  ([#31](https://github.com/mesosphere/kudo-cassandra-operator/pull/31))

## [3.11.5-0.1.2] - 2020-01-22

### Changed

- Make the KUDO Cassandra operator work with KUDO 0.10.0.
  ([#34](https://github.com/mesosphere/kudo-cassandra-operator/pull/34))
- Fix issue that prevented setting a custom volume storage class.
  ([#42](https://github.com/mesosphere/kudo-cassandra-operator/pull/42))

## [3.11.5-0.1.1] - 2019-12-12

### Added

- Upgrade Apache Cassandra to 3.11.5
  ([#29](https://github.com/mesosphere/kudo-cassandra-operator/pull/29))
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
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.5-0.1.2...HEAD
[3.11.5-0.1.2]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.5-0.1.1...v3.11.5-0.1.2
[3.11.5-0.1.1]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.4-0.1.0...v3.11.5-0.1.1
[3.11.4-0.1.0]:
  https://github.com/mesosphere/kudo-cassandra-operator/releases/tag/v3.11.4-0.1.0
