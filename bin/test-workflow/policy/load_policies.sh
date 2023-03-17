#!/usr/bin/env sh

set -eo

if [ "$CONJUR_APPLIANCE_URL" != "" ]; then
  echo y | conjur init -u $CONJUR_APPLIANCE_URL -a $CONJUR_ACCOUNT --self-signed --force
fi

# check for unset vars after checking for appliance url
set -u

conjur login -i admin -p $CONJUR_ADMIN_PASSWORD

readonly POLICY_DIR="/policy"

# NOTE: generated files are prefixed with the test app namespace to allow for parallel CI
set -- "$POLICY_DIR/generated/$TEST_APP_NAMESPACE_NAME.authenticator-policy.yml" \
  "$POLICY_DIR/generated/$TEST_APP_NAMESPACE_NAME.app-identities-policy.yml" \
  "$POLICY_DIR/generated/$TEST_APP_NAMESPACE_NAME.app-identities-policy-jwt.yml" \
  "$POLICY_DIR/app-policy.yml" \
  "$POLICY_DIR/generated/$TEST_APP_NAMESPACE_NAME.app-grants.yml"

for policy_file in "$@"; do
  echo "Loading policy $policy_file..."
  conjur policy load -b root -f $policy_file
done

# load secret values for each app
set -- "test-summon-init-app" \
  "test-summon-sidecar-app" \
  "test-summon-sidecar-jwt-app" \
  "test-secretless-app" \
  "test-secretless-jwt-app" \
  "test-secrets-provider-k8s-app" \
  "test-secrets-provider-k8s-jwt-app" \
  "test-secrets-provider-p2f-app" \
  "test-secrets-provider-p2f-injected-app" \
  "test-secrets-provider-rotation-app" \
  "test-secrets-provider-p2f-jwt-app" \
  "test-secrets-provider-standalone-app"

for app_name in "$@"; do
  echo "Loading secret values for $app_name"
  conjur variable set -i "$app_name-db/password" -v $DB_PASSWORD
  conjur variable set -i "$app_name-db/username" -v "test_app"

  case "${TEST_APP_DATABASE}" in
  postgres)
    PORT=5432
    PROTOCOL=postgresql
    ;;
  mysql)
    PORT=3306
    PROTOCOL=mysql
    ;;
  *)
    echo "Expected TEST_APP_DATABASE to be 'mysql' or 'postgres', got '${TEST_APP_DATABASE}'"
    exit 1
    ;;
  esac
  db_host="test-app-backend.$TEST_APP_NAMESPACE_NAME.svc.cluster.local"
  db_address="$db_host:$PORT"

  if [ "$app_name" = "test-secretless-app" ] || [ "$app_name" = "test-secretless-jwt-app" ]; then
    # Secretless doesn't require the full connection URL, just the host/port
    conjur variable set -i "$app_name-db/url" -v "$db_address"
    conjur variable set -i "$app_name-db/port" -v "$PORT"
    conjur variable set -i "$app_name-db/host" -v "$db_host"
  else
    # The authenticator sidecar injects the full pg connection string into the
    # app environment using Summon
    conjur variable set -i "$app_name-db/url" -v "$PROTOCOL://$db_address/test_app"
  fi

  if [ "$app_name" = "test-secrets-provider-rotation-app" ]; then
    # The rotation test uses an additional value which will be changed later
    conjur variable set -i "$app_name-db/counter" -v "0"
  fi

  # Add some secrets that can be used in demos
  conjur variable set -i "my-app-db/dev/password"  -v "dev-env-p@ssw0rd"
  conjur variable set -i "my-app-db/dev/url"       -v "https://dev.example.com:8080/api?list=false#anchor"
  conjur variable set -i "my-app-db/dev/username"  -v "dev-env-username"
  conjur variable set -i "my-app-db/dev/port"      -v "12345"
  conjur variable set -i "my-app-db/dev/cert-base64" -v "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURwRENDQW95Z0F3SUJBZ0lRRmZuTzFld2c1bUNoUi8rUlRWeDJQekFOQmdrcWhraUc5dzBCQVFzRkFEQVkKTVJZd0ZBWURWUVFERXcxamIyNXFkWEl0YjNOekxXTmhNQjRYRFRJeE1USXdOekUzTlRneE9Wb1hEVEl5TVRJdwpOekUzTlRneE9Wb3dHekVaTUJjR0ExVUVBeE1RWTI5dWFuVnlMbTE1YjNKbkxtTnZiVENDQVNJd0RRWUpLb1pJCmh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTHZ1SEIwMWo2YytnVVBTaW5Ub2czQ09FeWUrd1RxZ2VPYlgKK1YwSnNlQnhSTnpHTWlSeWljWXIxWG9MeWh1MXpYcHEzQ1JKbDdZaDc1TGtQWENCblo3UlBxYW1yZXFxMWYyZwp2ODJidnNxRkhKQ3Y2WlhLVlNTRTJQY2xEMDZOZjFCUzVoc1FhTjhlblJCRURnTzRqbUszUjBUajBuWGNXMjJBCmZjSlVxQ3MvZm9GVHNVOWxmYitLNDVSWlBGWjRJNnpCQ25QcHg3bzJXVjNPUnpqQ0s5M1lDb2U3OHVVWmNFaGwKK2ZIdGtnQ3N2T1pFa2EvYjRpUjJZRUJmNWlxLzdjSjU4OFpEdVZwL0FyWW9tNHB0dnc0R2srbEdxRytaSjg5eQo1UElBaEplS3NvS3N6Y0xyR0tvYndQdERMY3doYzhQaTJNS2FMZTJsUUhRU1J5cDF1djhDQXdFQUFhT0I1akNCCjR6QU9CZ05WSFE4QkFmOEVCQU1DQmFBd0hRWURWUjBsQkJZd0ZBWUlLd1lCQlFVSEF3RUdDQ3NHQVFVRkJ3TUMKTUF3R0ExVWRFd0VCL3dRQ01BQXdId1lEVlIwakJCZ3dGb0FVN3J3RmVJV0lhbmNtQkdoMGFjUjZPY0pBTHN3dwpnWUlHQTFVZEVRUjdNSG1DRUdOdmJtcDFjaTV0ZVc5eVp5NWpiMjJDQ21OdmJtcDFjaTF2YzNPQ0ZXTnZibXAxCmNpMXZjM011WTI5dWFuVnlMVzl6YzRJWlkyOXVhblZ5TFc5emN5NWpiMjVxZFhJdGIzTnpMbk4yWTRJblkyOXUKYW5WeUxXOXpjeTVqYjI1cWRYSXRiM056TG5OMll5NWpiSFZ6ZEdWeUxteHZZMkZzTUEwR0NTcUdTSWIzRFFFQgpDd1VBQTRJQkFRQzRjZEtWMGZ4NC96WkZwYUx1bmM0MGY0blkwUHZzRFRaZUNCM0w1M1c5M29hVGFjeTFEcWVFCnVScWxVT05yWWNsQitzOUYyTWVmRHBTK0swTHlVV21hVG9oM05JWGtKTDZHNS9TTHVuN01ZRXJwMXN0YVpDY3UKMkxQcjR1bDJMRmY3UVJjRytwRXIvTmxmay9RaENNb1NCYk1MNmZvYWtSQzNpTzA2bEUrMWk2MHBSdzdkdW90YgpoV3R0cUhqSTFMc1QxZ09neDhOVTNLeUczNERqNURGZFdQbU1FckdjWlNpMXNjdmlLL0UvMTkyUUxGbnB4UlR2CnltQ1J0eWpaUjFyaUM5ZkhqcUhuNGZ1bVV3eUJ5aEJpUmVmNXhGVHBiL2lQZzB6a0lIeENzd3ZQY3lod1djaHMKZGlYeElPUEVxWGNNN0xRbnNvalBoUEt0VXlJRytKdWwKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQoK"
  conjur variable set -i "my-app-db/prod/password" -v "prod-env-p@ssw0rd"
  conjur variable set -i "my-app-db/prod/url"      -v "https://prod.example.com:8080/api?list=false#anchor"
  conjur variable set -i "my-app-db/prod/username" -v "prod-env-username"
  conjur variable set -i "my-app-db/prod/port"     -v "12345"
  conjur variable set -i "my-app-db/prod/cert-base64" -v "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURwRENDQW95Z0F3SUJBZ0lRRmZuTzFld2c1bUNoUi8rUlRWeDJQekFOQmdrcWhraUc5dzBCQVFzRkFEQVkKTVJZd0ZBWURWUVFERXcxamIyNXFkWEl0YjNOekxXTmhNQjRYRFRJeE1USXdOekUzTlRneE9Wb1hEVEl5TVRJdwpOekUzTlRneE9Wb3dHekVaTUJjR0ExVUVBeE1RWTI5dWFuVnlMbTE1YjNKbkxtTnZiVENDQVNJd0RRWUpLb1pJCmh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTHZ1SEIwMWo2YytnVVBTaW5Ub2czQ09FeWUrd1RxZ2VPYlgKK1YwSnNlQnhSTnpHTWlSeWljWXIxWG9MeWh1MXpYcHEzQ1JKbDdZaDc1TGtQWENCblo3UlBxYW1yZXFxMWYyZwp2ODJidnNxRkhKQ3Y2WlhLVlNTRTJQY2xEMDZOZjFCUzVoc1FhTjhlblJCRURnTzRqbUszUjBUajBuWGNXMjJBCmZjSlVxQ3MvZm9GVHNVOWxmYitLNDVSWlBGWjRJNnpCQ25QcHg3bzJXVjNPUnpqQ0s5M1lDb2U3OHVVWmNFaGwKK2ZIdGtnQ3N2T1pFa2EvYjRpUjJZRUJmNWlxLzdjSjU4OFpEdVZwL0FyWW9tNHB0dnc0R2srbEdxRytaSjg5eQo1UElBaEplS3NvS3N6Y0xyR0tvYndQdERMY3doYzhQaTJNS2FMZTJsUUhRU1J5cDF1djhDQXdFQUFhT0I1akNCCjR6QU9CZ05WSFE4QkFmOEVCQU1DQmFBd0hRWURWUjBsQkJZd0ZBWUlLd1lCQlFVSEF3RUdDQ3NHQVFVRkJ3TUMKTUF3R0ExVWRFd0VCL3dRQ01BQXdId1lEVlIwakJCZ3dGb0FVN3J3RmVJV0lhbmNtQkdoMGFjUjZPY0pBTHN3dwpnWUlHQTFVZEVRUjdNSG1DRUdOdmJtcDFjaTV0ZVc5eVp5NWpiMjJDQ21OdmJtcDFjaTF2YzNPQ0ZXTnZibXAxCmNpMXZjM011WTI5dWFuVnlMVzl6YzRJWlkyOXVhblZ5TFc5emN5NWpiMjVxZFhJdGIzTnpMbk4yWTRJblkyOXUKYW5WeUxXOXpjeTVqYjI1cWRYSXRiM056TG5OMll5NWpiSFZ6ZEdWeUxteHZZMkZzTUEwR0NTcUdTSWIzRFFFQgpDd1VBQTRJQkFRQzRjZEtWMGZ4NC96WkZwYUx1bmM0MGY0blkwUHZzRFRaZUNCM0w1M1c5M29hVGFjeTFEcWVFCnVScWxVT05yWWNsQitzOUYyTWVmRHBTK0swTHlVV21hVG9oM05JWGtKTDZHNS9TTHVuN01ZRXJwMXN0YVpDY3UKMkxQcjR1bDJMRmY3UVJjRytwRXIvTmxmay9RaENNb1NCYk1MNmZvYWtSQzNpTzA2bEUrMWk2MHBSdzdkdW90YgpoV3R0cUhqSTFMc1QxZ09neDhOVTNLeUczNERqNURGZFdQbU1FckdjWlNpMXNjdmlLL0UvMTkyUUxGbnB4UlR2CnltQ1J0eWpaUjFyaUM5ZkhqcUhuNGZ1bVV3eUJ5aEJpUmVmNXhGVHBiL2lQZzB6a0lIeENzd3ZQY3lod1djaHMKZGlYeElPUEVxWGNNN0xRbnNvalBoUEt0VXlJRytKdWwKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQoK"

  echo "Adding JWT variables"
  conjur variable set -i conjur/authn-jwt/$AUTHENTICATOR_ID/jwks-uri -v $JWKS_URI
  conjur variable set -i conjur/authn-jwt/$AUTHENTICATOR_ID/token-app-property -v "sub"
  conjur variable set -i conjur/authn-jwt/$AUTHENTICATOR_ID/issuer -v $ISSUER
  conjur variable set -i conjur/authn-jwt/$AUTHENTICATOR_ID/identity-path -v "conjur/authn-jwt/$AUTHENTICATOR_ID/apps"
done

conjur logout
