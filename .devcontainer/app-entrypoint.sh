#!/usr/bin/env bash
set -euo pipefail

cd /workspace

# Git safe.directory (Windows bind mount에서 dubious ownership 방지)
if command -v git >/dev/null 2>&1; then
  git config --global --add safe.directory /workspace || true
fi

# VS Code server 캐시 권한 이슈 예방(EACCES)
mkdir -p /home/vscode/.cache/Microsoft || true
chown -R vscode:vscode /home/vscode/.cache || true

# go.sum/go.mod 정리(처음 1회는 시간이 걸릴 수 있음)
if [ -f go.mod ]; then
  go mod download || true
fi

if [ -f .air.toml ]; then
  exec air --config /workspace/.air.toml
fi

exec go run ./cmd/chatd