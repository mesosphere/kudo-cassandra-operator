# KUDO Cassandra Operator

The KUDO Cassandra Operator makes it easy to deploy and manage
[Apache Cassandra](http://cassandra.apache.org/) on Kubernetes.

| Konvoy                                                                                                                                                                                                                                                                                                                                                                                                      |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| <a href="https://teamcity.mesosphere.io/viewType.html?buildTypeId=Frameworks_DataServices_Kudo_Cassandra_Nightly_CassandraNightlyKonvoyKudo&branch_Frameworks_DataServices_Kudo_Cassandra_Nightly=%3Cdefault%3E&tab=buildTypeStatusDiv"><img src="https://teamcity.mesosphere.io/app/rest/builds/buildType:(id:Frameworks_DataServices_Kudo_Cassandra_Nightly_CassandraNightlyKonvoyKudo)/statusIcon"/></a> |

## Features

- Configurable `cassandra.yaml` and `jvm.options` parameters
- JVM memory locking out of the box
- Prometheus metrics and Grafana dashboard
- Horizontal scaling
- Rolling parameter updates
- Readiness probe
- Unpriviledged container execution

## Roadmap

- TLS
- Rack-awareness
- Node replace
- Backup/restore
- Inter-pod anti-affinity
- RBAC, pod security policies
- Liveness probe
- Multi-datacenter support
- Diagnostics bundle

## Getting started

## Documentation

## Version Chart

| Apache Cassandra version | Operator version | Minimum KUDO Version | Status |
| ------------------------ | ---------------- | -------------------- | ------ |
| 3.11.4                   | 0.1.0            | 0.8.0                | beta   |
