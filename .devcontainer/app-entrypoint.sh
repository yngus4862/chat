#!/usr/bin/env bash
set -euo pipefail

cd /workspace
echo "[devcontainer] app-entrypoint starting..."

# air가 없으면(설치 실패 등) 컨테이너를 살려둔 채로 디버깅 가능하게 함
if ! command -v air >/dev/null 2>&1; then
  echo "[devcontainer] ERROR: air not found in PATH"
  echo "[devcontainer] PATH=$PATH"
  tail -f /dev/null
fi

# .air.toml이 없으면 air가 즉시 종료하므로 컨테이너를 살려둠
if [ ! -f "/workspace/.air.toml" ]; then
  echo "[devcontainer] ERROR: /workspace/.air.toml not found"
  echo "[devcontainer] Listing /workspace:"
  ls -al /workspace | sed -n '1,120p'
  tail -f /dev/null
fi

echo "[devcontainer] launching: air -c /workspace/.air.toml"
set +e
air -c /workspace/.air.toml
rc=$?
set -e

echo "[devcontainer] air exited with code=$rc (keeping container alive for debugging)"
tail -f /dev/null