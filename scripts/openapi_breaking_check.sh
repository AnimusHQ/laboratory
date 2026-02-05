#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASELINE_DIR="${ROOT_DIR}/open/api/baseline"
CURRENT_DIR="${ROOT_DIR}/open/api/openapi"

if [ "${OPENAPI_BASELINE_UPDATE:-}" = "1" ]; then
  mkdir -p "${BASELINE_DIR}"
  cp "${CURRENT_DIR}"/*.yaml "${BASELINE_DIR}/"
  echo "openapi-compat: baseline updated"
  exit 0
fi

if [ "${OPENAPI_BREAKING_ALLOW:-}" = "1" ]; then
  echo "openapi-compat: OPENAPI_BREAKING_ALLOW=1; skipping."
  exit 0
fi

if [ ! -d "${BASELINE_DIR}" ]; then
  echo "openapi-compat: baseline directory missing: ${BASELINE_DIR}"
  exit 1
fi

changed=0

for baseline in "${BASELINE_DIR}"/*.yaml; do
  name="$(basename "${baseline}")"
  current="${CURRENT_DIR}/${name}"
  if [ ! -f "${current}" ]; then
    echo "openapi-compat: missing current spec ${name}"
    changed=1
    continue
  fi
  if ! diff -u "${baseline}" "${current}" >/dev/null; then
    echo "openapi-compat: spec drift detected for ${name}"
    changed=1
  fi
done

for current in "${CURRENT_DIR}"/*.yaml; do
  name="$(basename "${current}")"
  if [ ! -f "${BASELINE_DIR}/${name}" ]; then
    echo "openapi-compat: baseline missing for ${name}"
    changed=1
  fi
done

if [ "${changed}" -ne 0 ]; then
  echo "openapi-compat: baseline mismatch. Set OPENAPI_BASELINE_UPDATE=1 to refresh."
  exit 1
fi

echo "openapi-compat: baseline matches"
