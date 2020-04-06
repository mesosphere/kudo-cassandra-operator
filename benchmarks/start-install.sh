#!/usr/bin/env bash

devkudo="/Users/aneumann/go/src/github.com/kudobuilder/kudo/bin/kubectl-kudo"

$devkudo install ./operator --instance="cassandra" --namespace="cass-4dc" --parameter-file benchmarks/mwt/4dc/params-4dc.yaml
$devkudo install ./operator --instance="cassandra" --namespace="cass-3dc" --parameter-file benchmarks/mwt/3dc/params-3dc.yaml
$devkudo install ./operator --instance="cassandra" --namespace="cass-2dc" --parameter-file benchmarks/mwt/2dc/params-2dc.yaml
$devkudo install ./operator --instance="cassandra" --namespace="cass-2dc-par" --parameter-file benchmarks/mwt/2dc-par/params-2dc-par.yaml
$devkudo install ./operator --instance="cassandra" --namespace="cass-1dc-big" --parameter-file benchmarks/mwt/1dc-big/params-1dc-big.yaml
$devkudo install ./operator --instance="cassandra" --namespace="cass-1dc-small" --parameter-file benchmarks/mwt/1dc-small/params-1dc-small.yaml
$devkudo install ./operator --instance="cassandra" --namespace="cass-1dc-small-par" --parameter-file benchmarks/mwt/1dc-small-par/params-1dc-small-par.yaml
