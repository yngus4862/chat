FROM golang:1.26-trixie

RUN apt-get update \
  && apt-get install -y --no-install-recommends git curl ca-certificates bash \
  && rm -rf /var/lib/apt/lists/*

# vscode user (UID/GID 1000) for bind-mount permission sanity
RUN groupadd --gid 1000 vscode \
  && useradd --uid 1000 --gid 1000 -m vscode \
  && install -d -m 0755 -o vscode -g vscode /home/vscode/.cache/go-build \
  && install -d -m 0755 -o vscode -g vscode /home/vscode/go

ENV GOPATH=/home/vscode/go
ENV PATH=/home/vscode/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Tools (install to /usr/local/bin)
RUN GOBIN=/usr/local/bin /usr/local/go/bin/go install github.com/air-verse/air@latest \
  && GOBIN=/usr/local/bin /usr/local/go/bin/go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
  && GOBIN=/usr/local/bin /usr/local/go/bin/go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /workspace