#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=/dev/null
source "${ROOT_DIR}/scripts/go_env.sh"

CHARTS=(
  "${ROOT_DIR}/closed/deploy/helm/animus-datapilot"
  "${ROOT_DIR}/closed/deploy/helm/animus-dataplane"
)

ARGS=()
for chart in "${CHARTS[@]}"; do
  ARGS+=("--chart" "$chart")
done

if [[ "$#" -gt 0 ]]; then
  ARGS+=("$@")
fi

exec go run "${ROOT_DIR}/cmd/helm-images" "${ARGS[@]}"
