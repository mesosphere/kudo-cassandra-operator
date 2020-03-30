#!/usr/bin/env bash

kudo_cassandra_instance_name="cassandra"
kudo_cassandra_instance_namespace="mwt"

svc_endpoint="${kudo_cassandra_instance_name}-svc.${kudo_cassandra_instance_namespace}.svc.cluster.local"

workload_name="workload-local"
num_clients=10
keyspace_name="workload-local"
duration=1h
replication_string="strategy=NetworkTopologyStrategy, ring1=3, ring2=3"
threads=250
consistency="LOCAL_ONE"

workload_resource_file=workload-local.yaml