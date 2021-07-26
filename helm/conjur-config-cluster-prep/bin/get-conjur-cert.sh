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
 
# Save the users directory and execute from the directory in which this script exists
user_dir=$(pwd)
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
  echo "    -s                             Display the fingerprint but skip prompting"
  echo "                                   the user to acknowledge it is trusted."
  echo "    -u <Conjur appliance URL>      Conjur appliance URL (required)."
  echo "    -v                             Verify the certificate."
}

function main() {
  # Set defaults for command line options
  openssl_deployment="openssl-test"
  filepath=""
  internal_addr=false
  conjur_url=""
  skip_fingerprint_check=false
  verify=false
 
  # Process command line options
  local OPTIND
  while getopts ':f:hid:su:v' flag; do
    case "${flag}" in
      f) filepath=${OPTARG} ;;
      h) print_usage; exit 0 ;;
      i) internal_addr=true ;;
      d) openssl_deployment=${OPTARG} ;;
      s) skip_fingerprint_check=true ;;
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
    if [[ "${filepath:0:1}" = "/" ]]; then
      # Absolute filepath
      cert_filepath="$filepath"
    else
      # Relative filepath
      cert_filepath="$user_dir/$filepath"
    fi
    echo "Saving certificate to $cert_filepath"
  else
    # User's perspective file path
    echo "Saving certificate to default path of $(pwd)/$cert_filepath"
  fi

  domain_name="$(get_domain_name $conjur_url)"
  port="$(get_port $conjur_url)"

  echo "Retrieving SSL certificate for DNS domain $domain_name:$port"
  ssl_cmd="openssl s_client -showcerts -connect $domain_name:$port < /dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p'"

  # Read the certificate from either internal or external Conjur instance
  if [[ "$internal_addr" = false ]]; then
    echo "Assuming Conjur instance is outside of the Kubernetes cluster."

    status=0
    openssl_ver="$(openssl version)" || status="$?"
    if [ "$status" != 0 ] ; then
      echo "OpenSSL is required to gather the SSL certificate"
      exit 1
    fi
    echo "Using " "$openssl_ver"

    # Retrieve certificate
    cert="$(sh -c "$ssl_cmd")"

    # Write certificate to destination file
    echo "$cert" > "$cert_filepath"

    # Verify if desired
    if [ "$verify" = true ] ; then
      verify_certificate "$cert_filepath" "$domain_name" "$port"
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
      k8s_verify_certificate "$openssl_pod" "$cert_filepath" "$domain_name"
    fi

    # Delete the openssl test deployment
    if [ "$deployment_was_created" = true ]; then
        delete_openssl_deployment "$openssl_deployment"
        deployment_was_created=false
    fi
  fi
  verify_fingerprint
}

# This function will check the certificate fingerprint and
# ask the user if they trust the certificate
function verify_fingerprint() {
  echo -e "\n\nThe Conjur server's certificate SHA-1 fingerprint is:"
  # get the cert fingerprint
  openssl x509 -in $cert_filepath -noout -fingerprint

  if [[ "$skip_fingerprint_check" = false ]]; then

    echo "To verify this certificate, we recommend running "
    echo "the following command on the Conjur server:"
    echo "openssl x509 -fingerprint -noout -in ~conjur/etc/ssl/conjur.pem"
    echo "See https://github.com/cyberark/conjur-authn-k8s-client/blob/master/helm/kubernetes-cluster-prep/README.md"
    echo "for information on checking the fingerprint in Kubernetes"

    read -p "Trust this certificate? Y/N (Default: no): " trust
    if [[ !("$trust" =~ ^([yY][eE][sS]|[yY])$) ]]; then
      echo "You decided not to trust the certificate"
      echo "removing" $cert_filepath
      rm $cert_filepath
      exit 1
    fi
  fi

}

# This function will validate the Conjur SSL certificate using
# curl for an external (outside of Kubernetes cluster) conjur instance.
function verify_certificate() {
  cert_filepath="$1"
  domain_name="$2"
  port="$3"

  echo "Verifying the certificate"
  curl_cmd="curl --cacert $cert_filepath https://$domain_name:$port >/dev/null"
  status=0
  sh -c "$curl_cmd" || status="$?"
  if [ "$status" -eq 0 ]; then
    echo "certificate is verified!"
  else
    echo "certificate failed verification"
    exit 1
  fi
}

# Get the domain name and port from a URL (strips off protocol and endpoints)
function get_domain_and_port() {
    echo "$1" | sed -e 's|^[^/]*//||' -e 's|/.*$||'
}

function get_domain_name() {
    echo "$(get_domain_and_port $1)" | cut -d: -f1
}

function get_port() {
    echo "$(get_domain_and_port $1)" | grep : | cut -d: -f2
    if [[ ${PIPESTATUS[1]} -eq 1 ]]; then
      echo 443
    fi
}

function get_openssl_deployment() {
    openssl_deployment="$1"

    kubectl get pod -l "app=$openssl_deployment" -o jsonpath='{.items[*].metadata.name}'
}

function get_openssl_pod() {
    openssl_deployment="$1"

    kubectl get pod -l "app=$openssl_deployment" -o jsonpath='{.items[*].metadata.name}'
}

function ensure_openssl_pod_created() {
    openssl_deployment="$1"

    # Create a test deployment if it hasn't been created already
    existing_deployment="$(get_openssl_pod $openssl_deployment)"
    if [ -z "$existing_deployment" ]; then
        echo "Creating SSL deployment $openssl_deployment"
        kubectl create deployment "$openssl_deployment" \
            --image cyberark/conjur-k8s-cluster-test:edge
        # Remember that we need to clean up the deployment that we just created
        deployment_was_created=true
        # Wait for Pod to be ready
        echo "Waiting for OpenSSL test pod to be ready"
        # Some flakiness here - wait currently will fail if the resource doesn't exist yet
        # See https://github.com/kubernetes/kubernetes/issues/83242
        # TODO: Remove sleep after this is fixed in kubectl
        sleep 5
        # Wait for Pod to be ready
        kubectl wait --for=condition=ready pod -l "app=$openssl_deployment"
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
  domain_name="$3"

  echo "File path to copy: $cert_filepath"
  echo "Copying Conjur certificate to openssl pod"
  cert_filename="$(basename $cert_filepath)"
  kubectl cp "$cert_filepath" "$ssl_pod":"$cert_filename"

  # Test CA certificate with curl
  echo "Testing CA certificate with curl"
  curl_cmd="curl --cacert $cert_filename https://$domain_name"
  status=0
  (kubectl exec "$ssl_pod" -- sh -c "$curl_cmd" > /dev/null) || status="$?"
  [ "$status" -eq 0 ] && echo "certificate is verified!" || echo "certificate failed verification"

  # Remove the certificate
  kubectl exec "$ssl_pod" -- rm "$cert_filename"
}

function delete_openssl_deployment() {
    openssl_deployment="$1"
    kubectl delete deployment "$openssl_deployment"
}

main "$@"
