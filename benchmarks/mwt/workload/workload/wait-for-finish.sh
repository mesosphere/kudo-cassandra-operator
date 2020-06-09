#!/usr/bin/env bash
set -e

#workload_file=$1
#echo "Finish workload from $workload_file"
#source ${workload_file}

kudo_cassandra_instance_namespace="cassandra"
workload_name="cassandra-stress"
num_clients=2

echo "Waiting for pods to be finished"

pods=`kubectl get pods -n "${kudo_cassandra_instance_namespace}" | grep "${workload_name}" | awk '{print $1}'`

echo "Polling pods to be finished..."
while : ; do
    all_finished="true"
    finished_pods=0
    for pod in ${pods}; do
      echo -n "."
      pod_done=`kubectl logs --tail 20 "$pod" -n "${kudo_cassandra_instance_namespace}" | grep 'END\|RuntimeException' || echo "NotDone"`
      if [[ ${pod_done} != "NotDone" ]]; then
        finished_pods=$((finished_pods+1))
      fi
    done
    echo "Finished Pods: $finished_pods of $num_clients"
    [[ ${finished_pods} == ${num_clients} ]] && break
done
