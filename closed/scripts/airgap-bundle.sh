#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
usage: ./closed/scripts/airgap-bundle.sh --output <dir> [--values <file>] [--image-repo <repo>] [--tag <tag>] [--include-infra <0|1>]

Creates an air-gapped install bundle containing:
  - Helm chart packages for control-plane + dataplane
  - Docker image tarball (images.tar)
  - images.txt + SHA256SUMS

The script does NOT pull images from the network. All required images must already exist locally.
USAGE
  exit 2
}

OUTPUT_DIR=""
VALUES_FILE=""
IMAGE_REPO=""
TAG=""
INCLUDE_INFRA="1"

while [ $# -gt 0 ]; do
  case "$1" in
    --output)
      OUTPUT_DIR="${2:-}"
      shift 2
      ;;
    --values)
      VALUES_FILE="${2:-}"
      shift 2
      ;;
    --image-repo)
      IMAGE_REPO="${2:-}"
      shift 2
      ;;
    --tag)
      TAG="${2:-}"
      shift 2
      ;;
    --include-infra)
      INCLUDE_INFRA="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "unknown arg: $1" >&2
      usage
      ;;
  esac
done

if [ -z "${OUTPUT_DIR}" ]; then
  usage
fi

if ! command -v helm >/dev/null 2>&1; then
  echo "helm not found" >&2
  exit 2
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found (required to produce images.tar)" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHART_CP="${ROOT_DIR}/closed/deploy/helm/animus-datapilot"
CHART_DP="${ROOT_DIR}/closed/deploy/helm/animus-dataplane"

mkdir -p "${OUTPUT_DIR}"

VALUES_ARGS=()
TEMP_FILES=()

if [ -n "${VALUES_FILE}" ]; then
  VALUES_ARGS+=("--values" "${VALUES_FILE}")
fi

if [ -n "${IMAGE_REPO}" ] || [ -n "${TAG}" ]; then
  tmp_values="$(mktemp)"
  TEMP_FILES+=("${tmp_values}")
  cat > "${tmp_values}" <<EOF
image:
  repository: ${IMAGE_REPO:-animus}
  tag: ${TAG:-latest}
EOF
  VALUES_ARGS+=("--values" "${tmp_values}")
fi

if [ "${INCLUDE_INFRA}" != "1" ]; then
  tmp_infra="$(mktemp)"
  TEMP_FILES+=("${tmp_infra}")
  cat > "${tmp_infra}" <<EOF
postgres:
  enabled: false
minio:
  enabled: false
EOF
  VALUES_ARGS+=("--values" "${tmp_infra}")
fi

cleanup() {
  for file in "${TEMP_FILES[@]}"; do
    rm -f "${file}"
  done
}
trap cleanup EXIT

log() {
  echo "==> $*"
}

log "packaging Helm charts"
helm package "${CHART_CP}" --destination "${OUTPUT_DIR}" >/dev/null
helm package "${CHART_DP}" --destination "${OUTPUT_DIR}" >/dev/null

log "collecting image list"
"${ROOT_DIR}/scripts/list_images.sh" "${VALUES_ARGS[@]}" > "${OUTPUT_DIR}/images.txt"

if [ ! -s "${OUTPUT_DIR}/images.txt" ]; then
  echo "no images detected" >&2
  exit 2
fi

log "verifying images exist locally"
missing=()
while IFS= read -r img; do
  [ -z "${img}" ] && continue
  if ! docker image inspect "${img}" >/dev/null 2>&1; then
    missing+=("${img}")
  fi
done < "${OUTPUT_DIR}/images.txt"
if [ "${#missing[@]}" -gt 0 ]; then
  echo "missing local images:" >&2
  for img in "${missing[@]}"; do
    echo "  - ${img}" >&2
  done
  echo "build or load them first, then rerun the bundler." >&2
  exit 2
fi

log "writing images.tar"
mapfile -t images < "${OUTPUT_DIR}/images.txt"
docker save -o "${OUTPUT_DIR}/images.tar" "${images[@]}"

log "generating SHA256SUMS"
(cd "${OUTPUT_DIR}" && sha256sum *.tgz images.txt images.tar > SHA256SUMS)

cat > "${OUTPUT_DIR}/README.txt" <<EOF
Animus — Air-gapped bundle

Contents:
  - Helm charts (*.tgz)
  - images.tar (container images)
  - images.txt (image list)
  - SHA256SUMS

Next steps:
  1) Verify integrity:
       sha256sum -c SHA256SUMS
  2) Load images on your build host:
       docker load -i images.tar
  3) Push images to a registry reachable by your Kubernetes cluster,
     or load them into the cluster (kind/minikube/etc).
  4) Install charts with your pinned values (see docs/ops/airgapped-install.md).
EOF

log "bundle created at ${OUTPUT_DIR}"
