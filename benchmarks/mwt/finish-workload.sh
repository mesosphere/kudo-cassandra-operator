#!/usr/bin/env bash
set -e

workload_file=$1
echo "Finish workload from $workload_file"
source ${workload_file}

echo "Waiting for pods to be finished"

pods=`kubectl get pod -n "${kudo_cassandra_instance_namespace}" | grep "${workload_name}" | awk '{print $1}'`

echo "Polling pods to be finished..."
while : ; do
    all_finished="true"
    finished_pods=0
    for pod in ${pods}; do
      echo -n "."
      pod_done=`kubectl logs --tail 20 "$pod" -n "${kudo_cassandra_instance_namespace}" | grep 'END\|RuntimeException' || echo "NotDone"`
      if [[ $pod_done != "NotDone" ]]; then
        finished_pods=$((finished_pods+1))
      fi
    done
    echo "Finished Pods: $finished_pods of $num_clients"
    [[ ${finished_pods} == ${num_clients} ]] && break
done

echo "All Pods are finished, download artifacts..."
mkdir -p ${workload_name}
for pod in ${pods}; do
  echo "Download artifacts from $pod"
  kubectl logs "$pod" -n "${kudo_cassandra_instance_namespace}" > ${workload_name}/stdout_$pod.txt
  kubectl cp "${kudo_cassandra_instance_namespace}"/$pod:/tmp/${workload_name}.html ${workload_name}/workload_$pod.html
done

echo "Packaging artifacts..."
tar cvfz workload-a.tar.gz ${workload_name}

rm -r ${workload_name}

echo "Deleting workload resource..."
kubectl delete -f ${workload_resource_file}
