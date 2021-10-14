
set -e

source ./utils.sh

apply_deployment_patch app-test test-app-secrets-provider-p2f patch-custom-template.yaml
