#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="mwt"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-b"
num_clients=50
keyspace_name="workload-b"
duration=10m
replication_string="strategy=NetworkTopologyStrategy, ring1=3, ring2=3, ring3=3, ring4=3"
threads=500
consistency="LOCAL_QUORUM"


workload_resource_file=workload-b.yaml