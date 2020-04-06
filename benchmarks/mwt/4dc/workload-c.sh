#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cass-4dc"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-4dc-c"
num_clients=200
keyspace_name="workload-4dc-c2"
duration=10m
replication_string="strategy=NetworkTopologyStrategy, ring1=5, ring2=5, ring3=5, ring4=5"
threads=250
consistency="EACH_QUORUM"

workload_resource_file=workload-4dc-c.yaml