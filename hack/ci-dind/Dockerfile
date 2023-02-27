FROM golang:1.19.5-bullseye

LABEL maintainer=nitschneiderd@gmail.com

LABEL org.opencontainers.image.source=https://github.com/dnitsch/reststrategy/hack/ci-dind/Dockerfile
LABEL org.opencontainers.image.description="CI/CD helper for GHA for DinD"
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
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && apt-get update

RUN apt-get -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin

RUN apt-get clean

RUN groupadd -g 121 runner && useradd -mr -d /home/runner -u 1001 -g 121 runner \
  && usermod -aG sudo runner \
  && usermod -aG docker runner \
  && echo '%sudo ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

ENTRYPOINT [ "/bin/bash" ]
