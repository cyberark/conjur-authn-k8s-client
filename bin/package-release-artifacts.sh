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

  mkdir -p artifacts/raw-k8s-manifests
  cp conjur-config-cluster-prep/generated/* \
    conjur-config-namespace-prep/generated/* \
    artifacts/raw-k8s-manifests/

  rm -r conjur-config-*-prep/generated

  # helm charts
  for chart in "${HELM_CHARTS[@]}"; do
    helm package $chart
  done

  mkdir -p artifacts/helm-charts/
  mv *.tgz artifacts/helm-charts/
  cp conjur-config-cluster-prep/bin/get-conjur-cert.sh artifacts/

  # combine all the artifacts into a single archive file
  # use the repo name and version in the archive filename
  readonly REPO_NAME="${GITHUB_REPOSITORY##*/}"
  readonly TAG_NAME="${GITHUB_REF##*/}"
  readonly ARCHIVE_FILENAME="$REPO_NAME-$TAG_NAME.tar.gz"
  tar -czf $ARCHIVE_FILENAME artifacts

  rm -r artifacts/*
  mv $ARCHIVE_FILENAME artifacts/

  # cleanup
  for chart in "${HELM_CHARTS[@]}"; do
    git restore $chart/generated
  done
popd > /dev/null
