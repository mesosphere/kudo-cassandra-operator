#!/usr/bin/env bash

set -e

workload_file=$1
echo "Start workload from $workload_file"
source ${workload_file}

./kassandra-stress --kubernetes-namespace "${kudo_cassandra_instance_namespace}"   --workload-name "${workload_name}"   --hosts "${svc_endpoint}"   --number-of-clients ${num_clients} --keyspace-name ${keyspace_name} --duration "${duration}" --keyspace-replication-string="${replication_string}" --threads-per-client=${threads} --consistency-level=${consistency} --output-file=${workload_resource_file} --template-file=kassandra_stress.yaml.template

echo "Generated Workload Resources"

kubectl apply -f ${workload_resource_file}

echo "Applied Workload Resources"

echo "Waiting for all Workload Pods to be deployed"

while : ; do
    num_pods=`kubectl get pod -n "${kudo_cassandra_instance_namespace}" | grep "${workload_name}" | wc -l`
    [[ $num_pods == $num_clients ]] && break
    sleep 1
done

echo "All pods are deployed, use finish-workload.sh to continue"
