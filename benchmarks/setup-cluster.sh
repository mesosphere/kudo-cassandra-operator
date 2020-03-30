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

NAMESPACES="cassandra-dry"
#NAMESPACES="cassandra-4dc cassandra-2dc cassandra-1dc"

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

dev-kudo install ../../operator --instance="cassandra-dry" --namespace="cassandra-dry" --parameter-file dry/params-dry.yaml

#dev-kudo install ./operator --instance="cassandra-4dc" --namespace="cassandra-4dc" --parameter-file benchmarks/mwt/4dc/params-4dc.yaml
#dev-kudo install ./operator --instance="cassandra-2dc-a" --namespace="cassandra-2dc" --parameter-file benchmarks/mwt/2dc/params-2dc-a.yaml
#dev-kudo install ./operator --instance="cassandra-2dc-b" --namespace="cassandra-2dc" --parameter-file benchmarks/mwt/2dc/params-2dc-b.yaml

