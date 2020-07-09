#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/cloud-ark/labs/postgres-operator-sdk/pkg/generated \
github.com/cloud-ark/labs/postgres-operator-sdk/pkg/apis \
postgres-operator-sdk:v1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
