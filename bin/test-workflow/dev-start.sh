#!/usr/bin/env bash

# This script is intended for use by developers and not to be called by automation.

set -euo pipefail
IFS=$'\n\t'
export GCLOUD_PROJECT_NAME="${GCLOUD_PROJECT_NAME:-gke}"
export GCLOUD_ZONE="${GCLOUD_ZONE:-gke}"
export GCLOUD_CLUSTER_NAME="${GCLOUD_CLUSTER_NAME:-gke}"
export GCLOUD_SERVICE_KEY="${GCLOUD_SERVICE_KEY:-gke}"

print_usage() {
  echo "Usage:"
  echo "    This script will run start for the various platforms"
  echo ""
  echo "Syntax:"
  echo "    $0 [Options]"
  echo "    Options:"
  echo "    -c                         Run clients from a container"
  echo "    -e <Summon environment>    The environment used in"
  echo "                               secrets.yml. Defaults to dev"
  echo "    -g                         GKE"
  echo "    -h                         Show help"
  echo "    -l                         Just Log in to the platform"
  echo "    -o                         Openshift"
  echo "    -n                         No cleanup"
  echo "    -s <Step #>                Run one of the numbered scripts"
  echo "    -v <Openshift version>     The Openshift version"
  echo "                               defaults to current"
}

function main() {
  # Process command line options
  local OPTIND
  openshift_version="current"
  env="dev"
  oc_selected=false
  gke_selected=false
  no_cleanup=false
  local_container=false
  cmd=("")
  while getopts ':ce:ghlnos:v:' flag; do
    case "${flag}" in
      c) local_container=true ;;
      e) env=${OPTARG} ;;
      g) CONJUR_PLATFORM="gke" ; gke_selected=true ;;
      h) print_usage; exit 0 ;;
      l) cmd=("./platform_login.sh") ;;
      n) no_cleanup=true ;;
      o) oc_selected=true ;;
      s) cmd="./"$(ls | grep ${OPTARG}_) ;; 
      v) openshift_version=${OPTARG} ;;

  *) echo "Invalid argument -${OPTARG}" >&2; echo; print_usage ; exit 1;;
    esac
  done
  shift $((OPTIND-1))

  if [[ "$oc_selected" = true && "$gke_selected" = true ]]; then
    echo "Invalid arguments, cannot set -g and -o at the same time" >&2; print_usage ; exit 1;
  fi
  if [[ "$gke_selected" = true ]]; then
    cmd=("./start -p gke")
  elif [[ "$oc_selected" = true ]]; then
    if [[ $cmd == "" ]]; then
      cmd=("./start -p oc" )
    fi
    if [[ "$no_cleanup" = true ]]; then
      cmd+=" -n"
    fi
    echo "Openshift"
    echo "Running" "${cmd}"
    # some scripts need these set
    export CONJUR_PLATFORM="openshift"
    export APP_PLATFORM="openshift"
    export RUN_CLIENT_CONTAINER="$local_container"
    summon -e openshift -D ENV=$env -D VER=$openshift_version \
        sh -c "${cmd}"
  else
    echo "Unknown platform"
  fi

}

main "$@"

