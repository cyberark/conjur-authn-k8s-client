FROM google/cloud-sdk

RUN mkdir -p /src
WORKDIR /src

# Install Docker client
RUN apt-get update -y && \
    apt-get install -y apt-transport-https ca-certificates curl gnupg2 software-properties-common wget && \
    curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | apt-key add - && \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") $(lsb_release -cs) stable" && \
    apt-get update && \
    apt-get install -y docker-ce && \
    rm -rf /var/lib/apt/lists/*

# Install kubectl CLI
ARG KUBECTL_CLI_URL
RUN wget -O /usr/local/bin/kubectl ${KUBECTL_CLI_URL:-https://storage.googleapis.com/kubernetes-release/release/v1.7.6/bin/linux/amd64/kubectl} && \
    chmod +x /usr/local/bin/kubectl

# Install Helm CLI
ARG HELM_CLI_VERSION
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
RUN chmod 700 get_helm.sh
RUN ./get_helm.sh --no-sudo --version ${HELM_CLI_VERSION:-v3.5.2}
