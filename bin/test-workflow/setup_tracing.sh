#!/usr/bin/env bash

set -euo pipefail
source utils.sh

announce "Setting up Jaeger for tracing"

helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
$cli create namespace "${JAEGER_NAMESPACE_NAME}"
set_namespace "${JAEGER_NAMESPACE_NAME}"
helm install jaeger jaegertracing/jaeger \
    --set provisionDataStore.cassandra=false \
    --set provisionDataStore.elasticsearch=true \
    --set storage.type=elasticsearch \
    --set elasticsearch.replicas=1 \
    --set elasticsearch.minimumMasterNodes=1

wait_for_it 300 "has_resource 'app.kubernetes.io/instance=jaeger,app.kubernetes.io/component=collector' '$JAEGER_NAMESPACE_NAME'"
# TODO: Even when the pods are available, the collector refuses connections. We need to wait until the collector is actually ready to
# accept requests. Workaround is to run `kubectl delete pod test-app-secrets-provider-init-...` so the pod will be recreated after the
# collector is ready.
