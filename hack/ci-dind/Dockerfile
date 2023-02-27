FROM ubuntu:22.04

LABEL maintainer=nitschneiderd@gmail.com


LABEL org.opencontainers.image.source=https://github.com/dnitsch/reststrategy/hack/ci-dind/Dockerfile
LABEL org.opencontainers.image.description="RestStrategy controller from kubebuilder"
LABEL org.opencontainers.image.licence=MIT

RUN apt-get update && \
    apt-get -y install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    jq \
    git

RUN mkdir -m 0755 -p /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && apt-get update

RUN apt-get -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

RUN apt-get clean

ENTRYPOINT [ "/bin/bash" ]

