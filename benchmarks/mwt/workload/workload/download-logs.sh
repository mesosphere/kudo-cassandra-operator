#!/usr/bin/env bash
set -e

kudo_cassandra_instance_namespace="cassandra"
workload_name="cassandra-stress"
num_clients=$1

pods=`kubectl get pods -n "${kudo_cassandra_instance_namespace}" | grep "${workload_name}" | awk '{print $1}'`

echo "Download artifacts..."
mkdir -p ${workload_name}
for pod in ${pods}; do
  echo "Download artifacts from $pod"
  kubectl logs "$pod" -n "${kudo_cassandra_instance_namespace}" > ${workload_name}/stdout_$pod.txt
#  kubectl cp "${kudo_cassandra_instance_namespace}"/$pod:/tmp/${workload_name}.html ${workload_name}/workload_$pod.html
done

echo "Packaging artifacts..."
tar cvfz ${workload_name}.tar.gz ${workload_name}

rm -r ${workload_name}
