#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

require_bin() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required tool: $1" >&2
    exit 1
  fi
}

require_bin docker

GO_VERSION="${ANIMUS_GO_VERSION:-1.25}"
IMAGE_REPO="${ANIMUS_IMAGE_REPO:-animus}"
IMAGE_TAG="${ANIMUS_IMAGE_TAG:-}"
DOCKERFILE="${ROOT_DIR}/closed/deploy/docker/Dockerfile.service"
UI_DOCKERFILE="${ROOT_DIR}/closed/deploy/docker/Dockerfile.ui"

if [[ -z "$IMAGE_TAG" ]]; then
  if command -v git >/dev/null 2>&1 && git -C "$ROOT_DIR" rev-parse --short HEAD >/dev/null 2>&1; then
    IMAGE_TAG="local-$(git -C "$ROOT_DIR" rev-parse --short HEAD)"
  else
    IMAGE_TAG="local-$(date +%Y%m%d%H%M%S)"
  fi
fi

SERVICES=(
  gateway
  experiments
  dataset-registry
  quality
  lineage
  audit
  dataplane
)

for svc in "${SERVICES[@]}"; do
  echo "build-images: ${svc}"
  docker build \
    -f "$DOCKERFILE" \
    --build-arg GO_VERSION="$GO_VERSION" \
    --build-arg SERVICE="$svc" \
    -t "${IMAGE_REPO}/${svc}:${IMAGE_TAG}" \
    "$ROOT_DIR"
done

if [[ "${ANIMUS_BUILD_UI:-0}" == "1" ]]; then
  echo "build-images: ui"
  docker build \
    -f "$UI_DOCKERFILE" \
    -t "${IMAGE_REPO}/ui:${IMAGE_TAG}" \
    "$ROOT_DIR"
fi

echo "build-images: tag ${IMAGE_REPO}:$IMAGE_TAG"
