#!/bin/bash

source ./utils.sh

echo
announce "Displaying Annotations Defined for Application Pod"
kubectl get pods --selector=app=test-app-secrets-provider-p2f -o json | jq '.items[].metadata.annotations'
