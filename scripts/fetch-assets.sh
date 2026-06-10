#!/usr/bin/env bash
#
# Downloads the Freedoom asset set (BSD-3-Clause) into assets/freedoom/.
# Freedoom is a free, drop-in, Doom-compatible replacement for the original DOOM
# data, which is not freely licensed. See LICENSE for attribution.
#
# The game runs without these assets (it falls back to flat placeholder colours);
# this script is provided so the licensed source art can be fetched on demand.

set -euo pipefail

VERSION="${FREEDOOM_VERSION:-0.13.0}"
URL="https://github.com/freedoom/freedoom/releases/download/v${VERSION}/freedoom-${VERSION}.zip"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="${REPO_ROOT}/assets/freedoom"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Fetching Freedoom v${VERSION} ..."
echo "  ${URL}"

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi
if ! command -v unzip >/dev/null 2>&1; then
  echo "error: unzip is required" >&2
  exit 1
fi

curl -fL "${URL}" -o "${TMP}/freedoom.zip"
unzip -o -q "${TMP}/freedoom.zip" -d "${TMP}/extracted"

mkdir -p "${DEST}"
# Copy the WAD files; texture extraction beyond this is out of scope here.
find "${TMP}/extracted" -name '*.wad' -exec cp {} "${DEST}/" \;

echo "Freedoom assets written to ${DEST}"
echo "See LICENSE for license and attribution."
