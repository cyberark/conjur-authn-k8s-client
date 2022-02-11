#!/bin/bash

source ./utils.sh

APP_NAMESPACE="${APP_NAMESPACE:-app-test}"
echo
announce "Displaying secret files in Pod at location '$secrets_dir'"
display_secret_files "$APP_NAMESPACE" test-app-secrets-provider-p2f
