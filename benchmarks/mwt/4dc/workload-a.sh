#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="mwt"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-a"
num_clients=50
keyspace_name="workload-a"
duration=10m
replication_string="strategy=SimpleStrategy, replication_factor=0"
threads=500
consistency="LOCAL_ONE"

workload_resource_file=workload-a.yaml