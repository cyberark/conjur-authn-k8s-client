#!/usr/bin/env bash

source ./utils.sh

export PLATFORM_CTL="$cli"
export LOG_ROOT_DIR="temp"
LOG_DIR="$LOG_ROOT_DIR/k8s-logs"
LOG_ARCHIVE_PATH="$LOG_DIR-$(date "+%Y%m%d%H%M%S").tgz"

"./log_resources.sh" "logs" "$CONJUR_NAMESPACE_NAME" "$TEST_APP_NAMESPACE_NAME"

echo "Compressing cucumber logs to: $LOG_ARCHIVE_PATH"
tar -zcvf "$LOG_ARCHIVE_PATH" -C "$LOG_DIR" .
