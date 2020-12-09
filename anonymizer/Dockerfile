#
#
# Build
FROM python:3.8-slim AS builder

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    gcc \
    unzip \
    python-virtualenv

RUN mkdir -p /opt/bin

# Default versions
ARG GOOGLE_CLOUD_SDK_VERSION=309.0.0

# Google specific dependencies
RUN echo "GOOGLE_CLOUD_SDK_VERSION: ${GOOGLE_CLOUD_SDK_VERSION}" \
    && curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz \
    && tar zxvf google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz google-cloud-sdk \
    && mv google-cloud-sdk /opt/google-cloud-sdk \
    && /opt/google-cloud-sdk/bin/gcloud --version

COPY  scripts/anonymize-repo.sh /opt/bin/anonymize-repo


#
#
# Final
FROM python:3.8-slim

RUN apt-get update && apt-get install -y \
      ca-certificates \
      git \
      wget \
      openssh-client \
      zip \
      dnsutils \
      libnss-wrapper \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt /opt
ENV PATH=/opt/bin:/opt/google-cloud-sdk/bin:$PATH \
    HOME=/infra/.user

WORKDIR /infra
RUN mkdir -p  /infra/.user/.ssh
COPY config/ssh_config /etc/ssh/ssh_config

ENTRYPOINT ["/opt/bin/anonymize-repo"]
CMD bash
