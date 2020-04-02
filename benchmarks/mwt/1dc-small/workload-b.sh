#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cass-2dc"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-1dcbig-b"
num_clients=200
keyspace_name="workload-1dcbig-b"
duration=20m
replication_string="strategy=SimpleStrategy, replication_factor=3"
threads=250
consistency="QUORUM"

workload_resource_file=workload-1dcbig-b.yaml