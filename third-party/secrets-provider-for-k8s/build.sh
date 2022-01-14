  docker build \
    --build-arg TAG="secrets-rotation" \
    --tag "secrets-provider-for-k8s:secrets-rotation" \
    --target "secrets-provider" \
    .
