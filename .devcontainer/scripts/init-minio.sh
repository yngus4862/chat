# filename: .devcontainer/scripts/init-minio.sh
set -eu

mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

echo "[minio-init] create bucket: $MINIO_BUCKET_ATTACHMENTS"
mc mb --ignore-existing "local/$MINIO_BUCKET_ATTACHMENTS"

echo "[minio-init] done"
