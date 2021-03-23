#!/bin/bash

set -euo pipefail

# This script will retrieve the Conjur SSL cert and save it to a file.
#
# Usage:
#     get-conjur-cert.sh -u Conjur_Appliance_URL [Options]
#
# Example 1:
# To retrieve and verify a certificate from a Conjur instance
# that is external to the Kubernetes cluster:
#     get-conjur-cert.sh -v -u https://conjur.example.com
#
# Example 2:
# To retrieve and verify a certificate from a Conjur
# instance that is internal to the Kubernetes cluster:
#     get-conjur-cert.sh -v -i -u https://conjur.conjur-ns.svc.cluster.local
 
# Execute from the directory in which this script exists
cd "$(dirname "$0")"

# Default destination filepath for certificate file and default test
# deployment name for openssl testing.
DEFAULT_FILEPATH="../files/conjur-cert.pem"
DEFAULT_TEST_DEPLOYMENT="openssl-test"

# Keep track if we created an openssl test deployment. If we created one,
# we should clean it up when done.
deployment_was_created=false

print_usage() {
  echo "Usage:"
  echo "    This script will retrieve a Conjur SSL certificate based on"
  echo "    a URL for the Conjur instance, and save the certificate to"
  echo "    a file."
  echo ""
  echo "Syntax:"
  echo "    $0 -u <Conjur appliance URL> [Options]"
  echo "    Options:"
  echo "    -d <k8s test deployment name>  Kubernetes deployment name to use for"
  echo "                                   an openssl test pod. This only applies"
  echo "                                   if the '-i' command option is used. This"
  echo "                                   defaults to '$DEFAULT_TEST_DEPLOYMENT'."
  echo "    -f <destination filepath>      Destination file for writing certificate."
  echo "                                   If not set, certificate will be written"
  echo "                                   to '$DEFAULT_FILEPATH'."
  echo "    -h                             Show help"
  echo "    -i                             Conjur appliance URL is a Kubernetes"
  echo "                                   cluster internal address."
  echo "    -u <Conjur appliance URL>      Conjur appliance URL (required)"
  echo "    -v                             Verify the certificate"
}

function main() {
  # Set defaults for command line options
  openssl_deployment="openssl-test"
  filepath=""
  internal_addr=false
  conjur_url=""
  verify=false
 
  # Process command line options
  local OPTIND
  while getopts ':f:hid:u:v' flag; do
    case "${flag}" in
      f) filepath=${OPTARG} ;;
      h) print_usage; exit 0 ;;
      i) internal_addr=true ;;
      d) openssl_deployment=${OPTARG} ;;
      u) conjur_url=${OPTARG} ;;
      v) verify=true ;;
      *) echo "Invalid argument -${OPTARG}" >&2; echo; print_usage ; exit 1;;
    esac
  done
  shift $((OPTIND-1))
 
  # Check for required argument
  echo "Conjur URL: $conjur_url"
  if [[ -z "$conjur_url" ]]; then
    echo "Error: The Conjur Appliance URL argument is required" >&2
    echo
    print_usage
    exit 1
  fi

  # Set default path to destination certificate file
  cert_filepath="$DEFAULT_FILEPATH"

  # Process destination filepath argument
  if [[ -n "$filepath" ]]; then
    if [[ "$filepath" = "/*" ]]; then
      # Absolute filepath
      cert_filepath="$filepath"
    else
      # Relative filepath
      cert_filepath="$(pwd)/$filepath"
    fi
    echo "Saving certificate to $cert_filepath"
  else
    # User's perspective file path
    echo "Saving certificate to $(pwd)/$cert_filepath"
  fi

  domain_name="$(get_domain_name $conjur_url)"
  echo "Retrieving SSL certificate for DNS domain $domain_name"
  ssl_cmd="openssl s_client -showcerts -connect $domain_name:443 < /dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p'"

  # Read the certificate from either internal or external Conjur instance
  if [[ "$internal_addr" = false ]]; then
    echo "Assuming Conjur instance is outside of the Kubernetes cluster."

    # Retrieve certificate
    cert="$(sh -c "$ssl_cmd")"

    # Write certificate to destination file
    echo "$cert" > "$cert_filepath"

    # Verify if desired
    if [ "$verify" = true ] ; then
      verify_certificate "$cert_filepath" "$domain_name"
    fi

  else
    echo "Assuming Conjur instance is inside the Kubernetes cluster."

    # Retrieve certificate
    ensure_openssl_pod_created "$openssl_deployment"
    openssl_pod="$(get_openssl_pod $openssl_deployment)"
    cert="$(k8s_retrieve_certificate $openssl_pod "$ssl_cmd")"

    # Write certificate to destination file
    echo "$cert" > "$cert_filepath"

    # Verify if desired
    if [ "$verify" = true ] ; then
      k8s_verify_certificate "$openssl_pod" "$cert_filepath" "$conjur_url"
    fi

    # Delete the openssl test deployment
    if [ "$deployment_was_created" = true ]; then
        delete_openssl_deployment "$openssl_deployment"
        deployment_was_created=false
    fi
  fi

}

# This function will validate the Conjur SSL certificate using
# curl for an external (outside of Kubernetes cluster) conjur instance.
function verify_certificate() {
  cert_filepath="$1"
  domain_name="$2"

  echo "Verifying the certificate"
  curl_cmd="curl --cacert $cert_filepath https://$domain_name >/dev/null"
  status=0
  sh -c "$curl_cmd" || status="$?"
  if [ "$status" -eq 0 ]; then
    echo "certificate is verified!"
  else
    echo "certificate failed verification"
    exit 1
  fi
}

# Get the domain name from a URL (strips off protocol and endpoints)
function get_domain_name() {
    echo "$1" | sed -e 's|^[^/]*//||' -e 's|/.*$||'
}

function get_openssl_pod() {
    openssl_deployment="$1"

    kubectl get pod -l "run=$openssl_deployment" -o jsonpath='{.items[*].metadata.name}'
}

function ensure_openssl_pod_created() {
    openssl_deployment="$1"

    # Create a test deployment if it hasn't been created already
    openssl_pod="$(get_openssl_pod $openssl_deployment)"
    if [ -z "$openssl_pod" ]; then
        kubectl run "$openssl_deployment" \
           --image cyberark/conjur-cli:5 \
           --command sleep infinity
        # Remember that we need to clean up the deployment that we just created
        deployment_was_created=true
        # Wait for Pod to be ready
        kubectl wait --for=condition=ready pod -l "run=$openssl_deployment"
    fi
}

function k8s_retrieve_certificate() {
    ssl_pod="$1"
    ssl_cmd="$2"

    kubectl exec "$ssl_pod" -- sh -c "$ssl_cmd"
}

# This function will validate the Conjur SSL certificate using
# curl for an internal (inside the Kubernetes cluster) conjur instance.
function k8s_verify_certificate() {
  ssl_pod="$1"
  cert_filepath="$2"
  conjur_url="$3"

  echo "File path to copy: $cert_filepath"
  echo "Copying Conjur certificate to openssl pod"
  cert_filename="$(basename $cert_filepath)"
  kubectl cp "$cert_filepath" "$ssl_pod":"$cert_filename"

  # Test CA certificate with curl
  echo "Testing CA certificate with curl"
  curl_cmd="curl --cacert $cert_filename $conjur_url"
  status=0
  (kubectl exec "$ssl_pod" -- sh -c "$curl_cmd" > /dev/null) || status="$?"
  [ "$status" -eq 0 ] && echo "certificate is verified!" || echo "certificate failed verification"

  # Remove the certificate
  kubectl exec "$ssl_pod" -- rm "$cert_filename"
}

function delete_openssl_deployment() {
    openssl_deployment="$1"
    kubectl delete pod "$openssl_deployment"
}

main "$@"
