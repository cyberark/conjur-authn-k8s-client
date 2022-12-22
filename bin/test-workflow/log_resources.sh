#!/bin/bash
set -uo pipefail

# Usage:
# ./log_resources.sh [log prefix] [namespace N]...
#
# Note: This script expects several environment variables to be
# defined and exported, some of which are sensitive/secret values.
# It is for this that we recommend to always call this script using summon.
#
# Usage:
# ./log_resources.sh log_prefix namespace(s)
#
# Usage:
# ./log_resources.sh log_prefix namespace(s)
#
# log prefix  : Used to identify pod container logs from other logs (ex. this
#               script is used to log the same resources per different scenario,
#               this their logfiles should noticibly different).
# namespace(s): Variadic parameter; all pod containers within each of the given
#               namespaces will have their log contents written.

VERBOSE=8
LOG_LEVELS=( "Emergency" "Alert" "Critical" "Error" "Warning" "Notice" "Info" "Debug")
LOG_ROOT_DIR=${LOG_ROOT_DIR:-"temp"}
FEATURE_NAME=$1; shift;
# <timestamp>/<feature-name>/<scenario-name>/<namespace>_*.log
LOG_PREFIX="$(date "+%Y%m%d%M%S")/$(echo "${FEATURE_NAME// /-}" | tr '[:upper:]' '[:lower:]')"
LOG_DIR="$LOG_ROOT_DIR/k8s-logs/$(hostname)/$LOG_PREFIX"

function verbose() {
    if [ $# -lt 2 ]
    then
        echo "Error: expecting a verbose level and a message"
        exit 2
    fi
    
    local level=$1

    if [ -n "$VERBOSE" ]; then
        if (( "$level" <= "$VERBOSE" )); then
            printf "%9s %s\n" "[${LOG_LEVELS[$level-1]}]" "$2"
        fi
    fi
}

main() {
  # ensure log dir exists
  mkdir -p "$LOG_DIR"

  for ns in "$@"
  do
      echo "Begin logging all resources in namespace: $ns"
      get_all_resources_log_path="$LOG_DIR/${ns}_get-all-resources.log"
      echo -e "\tLog path: $get_all_resources_log_path"
      log_all_resources "$ns" "$get_all_resources_log_path"
      
      echo "Begin describing resources in namespace: $ns"
      describe_all_resources_log_path="$LOG_DIR/${ns}_describe-all-resources.log"
      echo -e "Log path: $describe_all_resources_log_path"
      describe_all_resources "$ns" "$describe_all_resources_log_path"
      
      echo "Begin logging all events in namespace: $ns" 
      all_events_log_path="$LOG_DIR/${ns}_get-all-events.log"
      echo -e "Log path: $all_events_log_path"
      log_all_events "$ns" "$all_events_log_path"

      # log all pod containers...
      echo "Begin logging all pod containers in namespace: $ns"
      log_all_pod_containers "$ns"
  done
}

log_all_resources() {
  local ns=$1
  local log_path=$2
  "$PLATFORM_CTL" get all -n "$ns" > "$log_path"
  remark_empty_log_file "$log_path"
}

describe_all_resources() {
  local ns=$1
  local log_path=$2
  "$PLATFORM_CTL" describe all -n "$ns" > "$log_path"
  remark_empty_log_file "$log_path"
}

log_all_events() {
  local ns=$1
  local log_path=$2
  "$PLATFORM_CTL" get events -n "$ns" > "$log_path"
  remark_empty_log_file "$log_path"
}

log_all_pod_containers(){
  local ns
  ns=$1
  
  pods=$("$PLATFORM_CTL" get pods --namespace="$ns" -o json | jq -r ".items[].metadata.name")

  if [ -z "$pods" ]; then
    echo -e "No pods found $ns in namespace."
    return
  fi

  echo -e "Found pods: $pods"

  for p in $pods; do
      log_pod_containers "$p" "$ns"
  done
}

function log_pod_containers() {
  local p
  local ns

  p=$1
  ns=$2

  if [ $# -lt 2 ]
  then
      verbose 4 "Expecting min 2 arguments (mandatory: pod name, namespace), $p, $ns, got $#"
      exit 2
  fi
  
  containers=$("$PLATFORM_CTL" get pod "$p" --namespace="$ns" -o json | jq -r ".spec.containers[]?.name")
  init_containers=$("$PLATFORM_CTL" get pod "$p" --namespace="$ns" -o json | jq -r ".spec.initContainers[]?.name")
  
  verbose 8 "Init Containers Found: $init_containers"
  verbose 8 "Containers Found: $containers"

  if [ -z "$init_containers" ]; then
      verbose 7 "No containers found in pod: $p"
  else
      verbose 7 "Begin logging init containers found in pod: $p"
      # ... and log each of them
      for c in $init_containers; do
          log_container "$p" "$c"
          # Sleep between each container log attempt
          sleep 1
      done
  fi

  if [ -z "$containers" ]; then
      verbose 7 "No containers found in pod: $p"
  else
      verbose 7 "Begin logging containers found in pod: $p"
      # ... and log each of them
      for c in $containers; do
          log_container "$p" "$c"
          # Sleep between each container log attempt
          sleep 1
      done
  fi
}

function log_container() {
    if [ $# -lt 1 ]
    then
        verbose 4 "Expecting min 2 arguments (mandatory: pod name, container name)"
        exit 2
    fi

    p=$1
    c=$2
    
    if [ -z "$c" ]; then
        verbose 7 "Container '$c' not provided, skipping!"
        return
    fi

    output_path="$LOG_DIR/${ns}_${p}_${c}.log"
    echo -e "Log path: $output_path"

    # Skip logging if container state is ContainerCreating
    state=$("$PLATFORM_CTL" get pod "$p" --namespace="$ns" -o json \
        | jq -r ".status.containerStatuses[] | select(.name == \"$c\") \
        | .state.waiting.reason" \
        || echo "")
    if [ "$state" == 'ContainerCreating' ] || [ "$state" == 'PodInitializing' ]; then
        msg="Container [$p/$c] is not a loggable state (current state: $state)"
        verbose 7 "$msg" 
        echo "$msg" >> "$output_path"
        return
    fi

    # log the container
    # $PLATFORM_CTL logs --prefix=true not supported in 3.11
    "$PLATFORM_CTL" logs "$p" --namespace="$ns" --timestamps -c "$c" >> "$output_path"

    remark_empty_log_file "$output_path"
}

remark_empty_log_file(){
  log_path=$1

  # if the file exists and is empty, tell us
  if [ ! -s "$log_path" ] ; then
    echo "Resource(s) produced no log output." >> "$log_path"
  fi
}

main "$@"
