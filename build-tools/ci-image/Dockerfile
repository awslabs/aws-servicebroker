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
  python-pip \
  python2.7 \
  python3 \
  git \
  wget \
  openssh-client \
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

RUN wget -q https://apb-oc.s3.amazonaws.com/apb-oc/oc-linux-64bit.tar.gz \
    && tar -xf oc-linux-64bit.tar.gz \
    && mv oc-linux-64bit/oc /usr/bin/oc \
    && chmod +x /usr/bin/oc \
    && rm -rf oc-linux-64bit.tar.gz oc-linux-64bit/

RUN wget "https://bootstrap.pypa.io/get-pip.py" -O /tmp/get-pip.py \
    && python3 /tmp/get-pip.py \
    && pip3 install --upgrade \
        taskcat \
        pygithub \
        boto3 \
        j2cli[yaml] \
    && python2 /tmp/get-pip.py \
    && pip2 install --upgrade \
        apb \
        boto3 \
        openshift \
        j2cli[yaml] \
    && rm -fr /var/lib/apt/lists/* /tmp/* /var/tmp/* /root/.cache/*

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

COPY analyse.py /usr/bin/analyse.py
COPY build-apbs.sh /usr/bin/build-apbs.sh
COPY build-broker.sh /usr/bin/build-broker.sh
COPY test-apbs.sh /usr/bin/test-apbs.sh
COPY test_apb.py /usr/bin/test_apb.py
COPY start.sh /start.sh
COPY install_aws_service_broker.sh /install_aws_service_broker.sh
COPY k8s-aws-service-broker.yaml.j2 /k8s-aws-service-broker.yaml.j2
COPY ca.yml /ca.yml

RUN chmod +x /usr/bin/analyse.py
RUN chmod +x /usr/bin/build-apbs.sh
RUN chmod +x /usr/bin/build-broker.sh
RUN chmod +x /usr/bin/test-apbs.sh
RUN chmod +x /usr/bin/test_apb.py
RUN chmod a+x /start.sh
RUN chmod +x /install_aws_service_broker.sh

VOLUME /var/lib/docker
