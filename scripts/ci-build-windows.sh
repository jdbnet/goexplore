#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck source=ci-common.sh
source scripts/ci-common.sh

require_version

OUTDIR="${OUTDIR:-dist}"
mkdir -p "${OUTDIR}" bin

if [ ! -f frontend/dist/index.html ]; then
  echo "Missing frontend/dist; run prepare or download the frontend-build artifact." >&2
  exit 1
fi

WIN_PACKAGES="gcc mingw-w64 nsis"

echo "==> Installing Windows cross-build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq ${WIN_PACKAGES}

apply_product_version
install_wails3_cli
ensure_build_assets
sync_app_icons
retry 3 go mod download

echo "==> Installing frontend deps (for Wails packaging tasks)..."
cd frontend
npm ci
cd ..

WIN_BUILD_FLAGS="-tags production -trimpath -buildvcs=false -ldflags \"-w -s -H windowsgui -X main.Version=v${VERSION}\""

echo "==> Building Windows amd64 installer..."
retry 2 wails3 task windows:package GOOS=windows ARCH=amd64 BUILD_FLAGS="${WIN_BUILD_FLAGS}"

# NSIS OutFile is ../../../bin/${INFO_PROJECTNAME}-${ARCH}-installer.exe (see build/windows/nsis/project.nsi).
WIN_INSTALLER=""
for candidate in \
  "bin/GoExplore-amd64-installer.exe" \
  "bin/goexplore-amd64-installer.exe" \
  "build/windows/nsis/goexplore-installer.exe"; do
  if [ -f "${candidate}" ]; then
    WIN_INSTALLER="${candidate}"
    break
  fi
done
if [ -z "${WIN_INSTALLER}" ]; then
  shopt -s nullglob
  matches=(bin/*-amd64-installer.exe)
  shopt -u nullglob
  if [ "${#matches[@]}" -ge 1 ]; then
    WIN_INSTALLER="${matches[0]}"
  fi
fi
if [ -z "${WIN_INSTALLER}" ] || [ ! -f "${WIN_INSTALLER}" ]; then
  echo "Windows installer not found (expected bin/*-amd64-installer.exe from NSIS)" >&2
  ls -la bin/ 2>/dev/null || true
  exit 1
fi
echo "==> Using Windows installer: ${WIN_INSTALLER}"
install -m 0755 "${WIN_INSTALLER}" "${OUTDIR}/goexplore-windows-installer.exe"
ls -lh "${OUTDIR}/goexplore-windows-installer.exe"
