#!/bin/bash

resource="$(kubectl get pod -n app-test -l app=test-app-secrets-provider-p2f -o name)"
pod_name="${resource##*/}"
kubectl logs -f -n app-test "$pod_name" -c cyberark-secrets-provider-for-k8s
