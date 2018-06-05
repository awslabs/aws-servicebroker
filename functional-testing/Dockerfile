FROM debian:stretch

RUN \
  DEBIAN_FRONTEND=noninteractive apt-get update -y && \
  DEBIAN_FRONTEND=noninteractive apt-get -yy -q --no-install-recommends install \
  iptables \
  ebtables \
  ethtool \
  ca-certificates \
  conntrack \
  socat \
  git \
  nfs-common \
  glusterfs-client \
  cifs-utils \
  apt-transport-https \
  ca-certificates \
  curl \
  gnupg2 \
  software-properties-common \
  bridge-utils \
  ipcalc \
  aufs-tools \
  sudo \
  wget \
  openssh-client \
  vim \
  net-tools \
  && DEBIAN_FRONTEND=noninteractive apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN \
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
  apt-key export "9DC8 5822 9FC7 DD38 854A E2D8 8D81 803C 0EBF CD88" | gpg - && \
  echo "deb [arch=amd64] https://download.docker.com/linux/debian stretch stable" >> \
    /etc/apt/sources.list.d/docker.list && \
  DEBIAN_FRONTEND=noninteractive apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get -yy -q --no-install-recommends install \
    docker-ce \
  && DEBIAN_FRONTEND=noninteractive apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.24.1/minikube-linux-amd64 && chmod +x minikube
ENV MINIKUBE_WANTUPDATENOTIFICATION=false
ENV MINIKUBE_WANTREPORTERRORPROMPT=false
ENV CHANGE_MINIKUBE_NONE_USER=true

COPY fake-systemctl.sh /usr/local/bin/systemctl

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.9.1/bin/linux/amd64/kubectl && \
  chmod a+x kubectl && \
  mv kubectl /usr/local/bin

RUN curl -sLO https://download.svcat.sh/cli/latest/linux/amd64/svcat && \
  chmod +x ./svcat && \
  mv ./svcat /usr/local/bin/ && \
  svcat install plugin

RUN curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get > get_helm.sh && \
  chmod +x get_helm.sh && \
  ./get_helm.sh --version v2.8.2

COPY start.sh /start.sh
COPY ca.yml /ca.yml
COPY aws-servicebroker /usr/bin/aws-servicebroker

RUN chmod a+x /start.sh
RUN chmod +x /usr/bin/aws-servicebroker

VOLUME /var/lib/docker
