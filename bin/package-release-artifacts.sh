#!/bin/bash

set -e

readonly HELM_CHARTS=("conjur-config-cluster-prep" "conjur-config-namespace-prep")
readonly REPO_ROOT="$(git rev-parse --show-toplevel)"
# Run from the repository root regardless of from where this script is invoked.
cd "$REPO_ROOT"

pushd helm > /dev/null
  # manifests for non-helm deployment
  # include version in manifest filename
  for chart in "${HELM_CHARTS[@]}"; do
    version=$(yq eval '.version' $chart/Chart.yaml)
    mv $chart/generated/$chart.yaml $chart/generated/$chart-$version.yaml
  done

  zip -v -j conjur-config-raw-k8s-manifests.zip \
    conjur-config-cluster-prep/generated/* \
    conjur-config-namespace-prep/generated/*

  rm -r conjur-config-*-prep/generated

  # helm charts
  for chart in "${HELM_CHARTS[@]}"; do
    helm package $chart
  done

  mkdir -p artifacts
  mv *.tgz *.zip -t artifacts
  cp conjur-config-cluster-prep/bin/get-conjur-cert.sh artifacts/

  # cleanup
  for chart in "${HELM_CHARTS[@]}"; do
    git restore $chart/generated
  done
popd > /dev/null
