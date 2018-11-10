#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

"${SCRIPT_ROOT}/vendor/k8s.io/code-generator/generate-groups.sh" all \
  github.com/travis-ci/trvs-operator/pkg/client \
  github.com/travis-ci/trvs-operator/pkg/apis \
  travisci:v1
