#!/bin/bash

set -eo pipefail

cd "$(git rev-parse --show-toplevel)"

pushd helm/conjur-config-cluster-prep > /dev/null
    mkdir -p generated

    helm template cluster-prep . \
        -f sample-values.yaml \
        --render-subchart-notes \
        --skip-tests | tee generated/conjur-config-cluster-prep.yaml
popd > /dev/null
