FROM google/cloud-sdk

# On CyberArk dev laptops, golang module dependencies are downloaded with a
# corporate proxy in the middle. For these connections to succeed we need to
# configure the proxy CA certificate in build containers.
#
# To allow this script to also work on non-CyberArk laptops where the CA
# certificate is not available, we copy the (potentially empty) directory
# and update container certificates based on that, rather than rely on the
# CA file itself.
ADD build_ca_certificate /usr/local/share/ca-certificates/
RUN update-ca-certificates

RUN mkdir -p /src
WORKDIR /src

# Install Docker client
RUN apt-get update -y && \
    apt-get install -y apt-transport-https ca-certificates curl gnupg2 software-properties-common wget jq && \
    curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | apt-key add - && \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") $(lsb_release -cs) stable" && \
    apt-get update && \
    apt-get install -y docker-ce && \
    rm -rf /var/lib/apt/lists/*

# Install kubectl CLI
ARG KUBECTL_VERSION
RUN curl -LO https://dl.k8s.io/release/v"${KUBECTL_VERSION:-1.21.3}"/bin/linux/amd64/kubectl && \
    mv kubectl /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl

# Install OpenShift oc CLI
ARG OPENSHIFT_CLI_URL
RUN wget -O oc.tar.gz "${OPENSHIFT_CLI_URL:-https://mirror.openshift.com/pub/openshift-v4/clients/ocp/stable/openshift-client-linux.tar.gz}" && \
    tar xvf oc.tar.gz  && \
    mv oc /usr/local/bin/oc && \
    rm -rf oc.tar.gz

# Install Helm CLI
ARG HELM_CLI_VERSION
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
RUN chmod 700 get_helm.sh
RUN ./get_helm.sh --no-sudo --version ${HELM_CLI_VERSION:-v3.5.2}
