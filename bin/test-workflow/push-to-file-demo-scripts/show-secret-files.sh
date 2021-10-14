#!/bin/bash

source ./utils.sh

echo
announce "Displaying secret files in Pod at location '$secrets_dir'"
display_secret_files app-test test-app-secrets-provider-p2f
