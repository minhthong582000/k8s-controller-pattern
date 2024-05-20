#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
SCRIPT_ROOT="${SCRIPT_DIR}/.."

controller-gen paths="${SCRIPT_ROOT}/pkg/apis/..." \
    crd:crdVersions=v1 \
    output:crd:artifacts:config=deploy/crds
