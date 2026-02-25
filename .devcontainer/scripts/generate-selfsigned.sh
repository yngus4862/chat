# filename: .devcontainer/scripts/generate-selfsigned.sh
set -eu
mkdir -p .devcontainer/certs

openssl req -x509 -nodes -newkey rsa:2048 \
  -keyout .devcontainer/certs/tls.key \
  -out .devcontainer/certs/tls.crt \
  -days 365 \
  -subj "/CN=chat.local"

echo "generated: .devcontainer/certs/tls.crt, .devcontainer/certs/tls.key"
