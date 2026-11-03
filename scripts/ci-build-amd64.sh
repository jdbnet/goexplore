#!/usr/bin/env bash
# Local / all-in-one release build (sequential). CI uses parallel jobs instead.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${VERSION:-}" ]; then
  echo "VERSION must be set" >&2
  exit 1
fi

LINUX_PACKAGES="gcc pkg-config libgtk-4-dev libwebkitgtk-6.0-dev mingw-w64 nsis"
echo "==> Installing build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq ${LINUX_PACKAGES}

bash scripts/ci-prepare.sh
bash scripts/ci-build-linux.sh
bash scripts/ci-build-windows.sh

echo "==> amd64 build complete (all platforms)."
