# syntax=docker/dockerfile:1.4
FROM golang:1.26-trixie

RUN apt-get update \
    && apt-get install -y --no-install-recommends git curl ca-certificates bash \
    && rm -rf /var/lib/apt/lists/*

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} -m ${USERNAME} \
    && install -d -m 0755 -o ${USERNAME} -g ${USERNAME} /home/${USERNAME}/.cache/go-build

# Use a user-owned GOPATH to avoid permission surprises
ENV GOPATH=/home/${USERNAME}/go
# Make sure both Go toolchain and user-installed go tools are always on PATH
ENV PATH=/home/${USERNAME}/go/bin:/usr/local/go/bin:${PATH}

RUN install -d -m 0755 -o ${USERNAME} -g ${USERNAME} /home/${USERNAME}/go

USER ${USERNAME}

# Install Go tooling as the vscode user (binaries go to $GOPATH/bin)
RUN go install github.com/cosmtrek/air@latest && \
    go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

USER root

# MinIO client
RUN curl -sLo /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc \
    && chmod +x /usr/local/bin/mc

WORKDIR /workspace
EXPOSE 8080 8081

USER ${USERNAME}
CMD ["bash"]