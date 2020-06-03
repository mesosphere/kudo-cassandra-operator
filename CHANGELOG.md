# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.11.6-0.2.0] - 2020-06-02

### Changed

- Bumped Cassandra Prometheus Exporter to 2.3.4.
  ([#56](https://github.com/mesosphere/kudo-cassandra-operator/pull/56))
- IP addresses are now managed using a custom bootstrap binary.
  ([#94](https://github.com/mesosphere/kudo-cassandra-operator/pull/94))
- Bumped Cassandra to 3.11.6.
  ([#116](https://github.com/mesosphere/kudo-cassandra-operator/pull/116))
- Minimum required KUDO version is now 0.13.0.
  ([#123](https://github.com/mesosphere/kudo-cassandra-operator/pull/123))

### Added

- TLS encryption for node-to-node and client-to-node connections.
  ([#31](https://github.com/mesosphere/kudo-cassandra-operator/pull/31))
- External service to allow access to Cassandra from outside the cluster.
  ([#46](https://github.com/mesosphere/kudo-cassandra-operator/pull/46))
- Multi-Datacenter configuration.
  ([#55](https://github.com/mesosphere/kudo-cassandra-operator/pull/55))
- Allow JMX rpc to be accessed from within the cluster.
  ([#58](https://github.com/mesosphere/kudo-cassandra-operator/pull/58))
- Ability to tune podManagementPolicy to enable parallel deploy.
  ([#72](https://github.com/mesosphere/kudo-cassandra-operator/pull/72))
- Ability to automatically install service account and roles.
  ([#71](https://github.com/mesosphere/kudo-cassandra-operator/pull/71))
- Liveness probe.
  ([#73](https://github.com/mesosphere/kudo-cassandra-operator/pull/73))
- Repair plan.
  ([#77](https://github.com/mesosphere/kudo-cassandra-operator/pull/77))
- Nodetool SSL access via JMX.
  ([#74](https://github.com/mesosphere/kudo-cassandra-operator/pull/74))
- Password authentication.
  ([#88](https://github.com/mesosphere/kudo-cassandra-operator/pull/88))
- Backup to S3 using medusa.
  ([#60](https://github.com/mesosphere/kudo-cassandra-operator/pull/60),
  [#124](https://github.com/mesosphere/kudo-cassandra-operator/pull/124))
- Support for Cassandra clusters spanning multiple Kubernetes clusters.
  ([#97](https://github.com/mesosphere/kudo-cassandra-operator/pull/97))
- Support for custom prometheus exporter configuration.
  ([#93](https://github.com/mesosphere/kudo-cassandra-operator/pull/93))
- Ability to start new Cassandra nodes when a Kubernetes Cluster node fails via
  custom recovery controller.
  ([#96](https://github.com/mesosphere/kudo-cassandra-operator/pull/96))

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
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.6-0.2.0...HEAD
[3.11.6-0.2.0]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.5-0.1.2...v3.11.6-0.2.0
[3.11.5-0.1.2]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.5-0.1.1...v3.11.5-0.1.2
[3.11.5-0.1.1]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.4-0.1.0...v3.11.5-0.1.1
[3.11.4-0.1.0]:
  https://github.com/mesosphere/kudo-cassandra-operator/releases/tag/v3.11.4-0.1.0
