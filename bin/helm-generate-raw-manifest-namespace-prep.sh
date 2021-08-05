#!/bin/bash

set -eo pipefail

cd "$(git rev-parse --show-toplevel)"

pushd helm/conjur-config-namespace-prep > /dev/null
    mkdir -p generated

    helm template namespace-prep . \
        -f sample-values.yaml \
        --render-subchart-notes \
        --skip-tests | tee generated/conjur-config-namespace-prep.yaml
popd > /dev/null
