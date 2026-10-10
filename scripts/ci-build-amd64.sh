#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${VERSION:-}" ]; then
  echo "VERSION must be set" >&2
  exit 1
fi

OUTDIR="${OUTDIR:-dist}"
mkdir -p "${OUTDIR}"

LINUX_PACKAGES="gcc pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev mingw-w64 nsis"

echo "==> Installing build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq ${LINUX_PACKAGES}

echo "==> Installing Wails v2 CLI..."
CGO_ENABLED=0 go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
export PATH="$(go env GOPATH)/bin:${PATH}"

echo "==> Setting product version..."
sed -i "s/\"productVersion\": \".*\"/\"productVersion\": \"${VERSION}\"/" wails.json

echo "==> Building Linux amd64..."
wails build -platform linux/amd64 -tags webkit2_41 -clean -ldflags "-X main.Version=v${VERSION}"
install -m 0755 build/bin/goexplore "${OUTDIR}/goexplore-linux-amd64"

echo "==> Building Windows amd64 installer..."
wails build -platform windows/amd64 -clean -nsis -ldflags "-X main.Version=v${VERSION}"
install -m 0755 build/bin/GoExplore-amd64-installer.exe "${OUTDIR}/goexplore-windows-installer.exe"

echo "==> Building Debian package..."
bash scripts/ci-deb.sh amd64

echo "==> amd64 build complete:"
ls -lh "${OUTDIR}/goexplore-linux-amd64" \
  "${OUTDIR}/goexplore-windows-installer.exe" \
  "${OUTDIR}/goexplore_${VERSION}_amd64.deb"
