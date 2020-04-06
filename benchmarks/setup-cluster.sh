#!/usr/bin/env bash

echo "Create ClusterRole for node resolver"

kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: node-resolver-role
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "watch", "list"]
EOF

echo "Create Namespaces"

#NAMESPACES="cassandra-dry"
NAMESPACES="cass-4dc cass-3dc cass-2dc cass-2dc-par cass-1dc-big cass-1dc-small cass-1dc-small-par"

for NS in ${NAMESPACES}; do
    kubectl create namespace "${NS}"
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: node-resolver
  namespace: ${NS}
EOF
    kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: node-resolver-binding
subjects:
- kind: ServiceAccount
  name: node-resolver
  namespace: ${NS}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: node-resolver-role
EOF
done

#dev-kudo install ../../operator --instance="cassandra-dry" --namespace="cassandra-dry" --parameter-file dry/params-dry.yaml


