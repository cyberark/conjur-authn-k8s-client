#!/usr/bin/env bash
# Runs dlv for one debug only
/go/bin/dlv debug cmd/authenticator/main.go  \
    --listen=:40000 \
    --headless=true \
    --api-version=2 \
    --accept-multiclient
# Rerun again the script
bash dev/start.sh