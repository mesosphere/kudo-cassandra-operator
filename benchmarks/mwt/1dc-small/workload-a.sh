#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cass-1dc-small"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-1dcsmall-a"
num_clients=100
keyspace_name="workload-1dcsmall-a"
duration=15m
replication_string="strategy=SimpleStrategy, replication_factor=0"
threads=250
consistency="ONE"

workload_resource_file=workload-1dcsmall-a.yaml