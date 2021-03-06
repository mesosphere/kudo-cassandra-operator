# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.11.7-1.0.2] - 2020-11-04

### Changed

- Started extending Parameters with additional attributes and types
  ([#162](https://github.com/mesosphere/kudo-cassandra-operator/pull/162))
- Bumped Cassandra to 3.11.7
  ([#161](https://github.com/mesosphere/kudo-cassandra-operator/pull/161))

### Added

- Introduce a parameter to set annotations for the external service
  ([#164](https://github.com/mesosphere/kudo-cassandra-operator/pull/164))
- Add OpenSSL and zlib to Medusa Docker images
  ([#163](https://github.com/mesosphere/kudo-cassandra-operator/pull/163))

## [3.11.6-1.0.1] - 2020-07-24

### Changed

- Fix warnings when running 'kubectl kudo package verify'
  ([#157](https://github.com/mesosphere/kudo-cassandra-operator/pull/157))
- Use Toggle Task for deployment of Cassandra exporter
  ([#140](https://github.com/mesosphere/kudo-cassandra-operator/pull/140))
- Rework Readiness and Liveness Probes
  ([#155](https://github.com/mesosphere/kudo-cassandra-operator/pull/155))

### Added

- Add architecture/repair/decommission docs
  ([#154](https://github.com/mesosphere/kudo-cassandra-operator/pull/154))
- Create simple Operator for Workload generation
  ([#151](https://github.com/mesosphere/kudo-cassandra-operator/pull/151))
- Allow definition of node tolerations for tainted k8s nodes
  ([#153](https://github.com/mesosphere/kudo-cassandra-operator/pull/153))
- Add resources and production docs
  ([#126](https://github.com/mesosphere/kudo-cassandra-operator/pull/126))
- Add v1 dashboard
  ([#141](https://github.com/mesosphere/kudo-cassandra-operator/pull/141))

## [3.11.6-1.0.0] - 2020-06-04

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
- Backup and restore to/from S3 using medusa.
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
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.6-1.0.0...HEAD
[3.11.6-1.0.0]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.5-0.1.2...v3.11.6-1.0.0
[3.11.5-0.1.2]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.5-0.1.1...v3.11.5-0.1.2
[3.11.5-0.1.1]:
  https://github.com/mesosphere/kudo-cassandra-operator/compare/v3.11.4-0.1.0...v3.11.5-0.1.1
[3.11.4-0.1.0]:
  https://github.com/mesosphere/kudo-cassandra-operator/releases/tag/v3.11.4-0.1.0
