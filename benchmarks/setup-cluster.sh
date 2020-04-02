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

dev-kudo install ./operator --instance="cassandra" --namespace="cass-4dc" --parameter-file benchmarks/mwt/4dc/params-4dc.yaml
dev-kudo install ./operator --instance="cassandra" --namespace="cass-3dc" --parameter-file benchmarks/mwt/3dc/params-3dc.yaml
dev-kudo install ./operator --instance="cassandra" --namespace="cass-2dc" --parameter-file benchmarks/mwt/2dc/params-2dc.yaml
dev-kudo install ./operator --instance="cassandra" --namespace="cass-2dc-par" --parameter-file benchmarks/mwt/2dc-par/params-2dc-par.yaml
dev-kudo install ./operator --instance="cassandra" --namespace="cass-1dc-big" --parameter-file benchmarks/mwt/1dc-big/params-1dc-big.yaml
dev-kudo install ./operator --instance="cassandra" --namespace="cass-1dc-small" --parameter-file benchmarks/mwt/1dc-small/params-1dc-small.yaml
dev-kudo install ./operator --instance="cassandra" --namespace="cass-1dc-small-par" --parameter-file benchmarks/mwt/1dc-small-par/params-1dc-small-par.yaml

