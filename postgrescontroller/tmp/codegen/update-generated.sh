#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/demo/postgrescontroller/pkg/generated \
github.com/demo/postgrescontroller/pkg/apis \
postgrescontroller:v1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
