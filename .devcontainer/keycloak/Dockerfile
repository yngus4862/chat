# filename: .devcontainer/keycloak/Dockerfile
FROM keycloak/keycloak:26.5.0-0

USER root
RUN set -eux; \
  if command -v microdnf >/dev/null 2>&1; then \
    microdnf install -y curl ca-certificates && microdnf clean all; \
  elif command -v dnf >/dev/null 2>&1; then \
    dnf -y install curl ca-certificates && dnf clean all; \
  elif command -v yum >/dev/null 2>&1; then \
    yum -y install curl ca-certificates && yum clean all; \
  elif command -v apt-get >/dev/null 2>&1; then \
    apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*; \
  elif command -v apk >/dev/null 2>&1; then \
    apk add --no-cache curl ca-certificates; \
  else \
    echo "No supported package manager found (microdnf/dnf/yum/apt-get/apk)."; \
    exit 1; \
  fi

USER 1000
