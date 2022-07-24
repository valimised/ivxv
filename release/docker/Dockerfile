# IVXV Internet voting framework
# Dockerfile for release builder

# Ubuntu 20.04 LTS (Focal Fossa)
FROM ubuntu:focal
LABEL maintainer="IVXV Developer, info@ivotingcentre.ee"

# Set up debconf frontend
ENV DEBIAN_FRONTEND=noninteractive

# Set up APT
COPY apt-sources.list /etc/apt/sources.list
COPY apt-local-options /etc/apt/apt.conf.d/90local-options

# Install packages
COPY install-builder-packages.sh /usr/local/sbin/install-builder-packages.sh
RUN sh -x /usr/local/sbin/install-builder-packages.sh

# Install documentation tools
COPY requirements.txt /usr/src/requirements.txt
RUN pip3 install --requirement /usr/src/requirements.txt

# Install static analysis tools.
# Usually curl | sh is unsafe, but it should be okay inside Docker.
ARG golangcilint=v1.39.0
RUN curl --silent --show-error --fail --location \
        https://raw.githubusercontent.com/golangci/golangci-lint/$golangcilint/install.sh | \
        sh -s -- -b /usr/local/bin $golangcilint && \
        golangci-lint --version

# Add user for jenkins
RUN useradd --create-home jenkins

# Set up Estonian locale
ENV LANG=et_EE.UTF-8 LC_ALL=et_EE.UTF-8
