#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cass-2dc"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-2dc-b"
num_clients=100
keyspace_name="workload-2dc-b"
duration=20m
replication_string="strategy=NetworkTopologyStrategy, ring1=3, ring2=3"
threads=250
consistency="LOCAL_QUORUM"

workload_resource_file=workload-2dc-b.yaml