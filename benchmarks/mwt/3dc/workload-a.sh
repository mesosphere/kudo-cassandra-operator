#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="cass-3dc"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-3dc-a"
num_clients=200
keyspace_name="workload-3dc-a"
duration=15m
replication_string="strategy=NetworkTopologyStrategy, ring1=1, ring2=1, ring3=1"
threads=250
consistency="LOCAL_ONE"

workload_resource_file=workload-3dc-a.yaml