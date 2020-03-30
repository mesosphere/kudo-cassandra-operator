#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra-dry"
kudo_cassandra_instance_namespace="cassandra-dry"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-dry"
num_clients=10
keyspace_name="workload-dry"
duration=5m
replication_string="strategy=NetworkTopologyStrategy, ring1=1"
threads=100
consistency="ONE"

workload_resource_file=workload-dry.yaml