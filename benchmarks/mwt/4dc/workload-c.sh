#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cass-4dc"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-4dc-c"
num_clients=400
keyspace_name="workload-4dc-c"
duration=20m
replication_string="strategy=NetworkTopologyStrategy, ring1=3, ring2=3, ring3=3, ring4=3"
threads=250
consistency="EACH_QUORUM"

workload_resource_file=workload-4dc-c.yaml