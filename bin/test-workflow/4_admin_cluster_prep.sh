#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")" || ( echo "cannot cd into dir" && exit 1 )

TIMEOUT="${TIMEOUT:-5m0s}"

source utils.sh

check_env_var CONJUR_APPLIANCE_URL
check_env_var CONJUR_NAMESPACE_NAME
check_env_var CONJUR_ACCOUNT
check_env_var AUTHENTICATOR_ID

if [[ "$CONJUR_OSS_HELM_INSTALLED" == "false" ]]; then
  check_env_var CONJUR_FOLLOWER_URL
fi

# Upon error, dump kubernetes resources in the Conjur Namespace
trap dump_conjur_namespace_upon_error EXIT

set_namespace default

# Prepare our cluster with conjur and authnK8s credentials in a golden configmap
announce "Installing cluster prep chart"
pushd ../../helm/conjur-config-cluster-prep > /dev/null

  if [[ "$CONJUR_OSS_HELM_INSTALLED" == "true" ]]; then
    conjur_url="$CONJUR_APPLIANCE_URL"
    get_cert_options="-v -i -s -u"
    additional_options="--set authnK8s.serviceAccount.create=false --set authnK8s.serviceAccount.name=conjur-oss"
  else
    conjur_url="$CONJUR_FOLLOWER_URL"
    if [[ "$CONJUR_PLATFORM" == "gke" ]]; then
      get_cert_options="-v -i -s -u"
      additional_options="--set authnK8s.serviceAccount.create=false --set authnK8s.serviceAccount.name=conjur-cluster"
    elif [[ "$CONJUR_PLATFORM" == "jenkins" ]]; then
      get_cert_options="-v -s -u"
      additional_options=""
    fi
  fi

  ./bin/get-conjur-cert.sh $get_cert_options "$conjur_url"
  helm upgrade --install "cluster-prep-$UNIQUE_TEST_ID" . -n "$CONJUR_NAMESPACE_NAME" --wait --timeout "$TIMEOUT" \
      --create-namespace \
      --set conjur.account="$CONJUR_ACCOUNT" \
      --set conjur.applianceUrl="$conjur_url" \
      --set conjur.certificateFilePath="files/conjur-cert.pem" \
      --set authnK8s.authenticatorID="$AUTHENTICATOR_ID" \
      --set authnK8s.clusterRole.name="conjur-clusterrole-$UNIQUE_TEST_ID" \
      $additional_options

popd > /dev/null

pushd temp > /dev/null
# Prepare our cluster with a sidecar injector
announce "Installing sidecar injector"
# Check for the secretless CRD and delete if found
check_for_crd
# Clean out old sidecar injector webhooks
clean_web_hooks
pushd "sidecar-injector-$UNIQUE_TEST_ID/helm/cyberark-sidecar-injector" > /dev/null
# Rename the chart to a shorted name due to an SSL overflow issue with long
# [service].[namespace] names
sed -i 's/cyberark-sidecar-injector/sidecar-injector/g' Chart.yaml
pushd templates

cat > serviceaccountsecret.yaml << EOF
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  name: {{ include "cyberark-sidecar-injector.name" . }}-service-account-token
  labels:
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
  annotations:
    kubernetes.io/service-account.name: {{ include "cyberark-sidecar-injector.name" . }}
EOF

popd > /dev/null
set_namespace $CONJUR_NAMESPACE_NAME

helm --namespace $CONJUR_NAMESPACE_NAME install cyberark-sidecar-injector \
      --set "deploymentApiVersion=apps/v1" \
      --set "sidecarInjectorImage=cyberark/sidecar-injector:edge" \
      --set "caBundle=$(kubectl -n kube-system get configmap extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' )"  .

popd > /dev/null
popd > /dev/null

announce "Approving Certificate"
"$cli" get deployment

wait_for_it 300 "$cli -n $CONJUR_NAMESPACE_NAME logs deployment/sidecar-injector -c init-webhook | grep 'waiting for certificate'"
announce "Done waiting"
"$cli" -n "$CONJUR_NAMESPACE_NAME" logs deployment/sidecar-injector -c init-webhook


kubectl certificate approve "sidecar-injector.$CONJUR_NAMESPACE_NAME"
announce "Approve"
sleep 3
"$cli" -n "$CONJUR_NAMESPACE_NAME" logs deployment/sidecar-injector -c init-webhook
announce "Done"
