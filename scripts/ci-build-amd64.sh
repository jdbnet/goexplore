#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${VERSION:-}" ]; then
  echo "VERSION must be set" >&2
  exit 1
fi

OUTDIR="${OUTDIR:-dist}"
mkdir -p "${OUTDIR}" bin

LINUX_PACKAGES="gcc pkg-config libgtk-4-dev libwebkitgtk-6.0-dev mingw-w64 nsis"

echo "==> Installing build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq ${LINUX_PACKAGES}

echo "==> Installing Wails v3 CLI..."
WAILS_VERSION="$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)"
CGO_ENABLED=0 go install "github.com/wailsapp/wails/v3/cmd/wails3@${WAILS_VERSION}"
export PATH="$(go env GOPATH)/bin:${PATH}"

echo "==> Setting product version..."
sed -i "s/\"productVersion\": \".*\"/\"productVersion\": \"${VERSION}\"/" wails.json

if [ ! -f Taskfile.yml ]; then
  echo "==> Generating Wails v3 build assets..."
  wails3 update build-assets -name "GoExplore" -binaryname goexplore -config ./wails.json -dir build
fi

echo "==> Generating frontend bindings..."
wails3 generate bindings -clean=true

echo "==> Building frontend..."
cd frontend
npm ci
npm run build
cd ..

LDFLAGS="-w -s -X main.Version=v${VERSION}"

echo "==> Building Linux amd64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags production -trimpath -ldflags "${LDFLAGS}" -o "${OUTDIR}/goexplore-linux-amd64" .

echo "==> Building Windows amd64 installer..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc wails3 task windows:package LD_FLAGS="-X main.Version=v${VERSION}" || {
  echo "==> Falling back to Windows executable only..."
  CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -tags production -trimpath -ldflags "-w -s -H windowsgui ${LDFLAGS#-w -s }" -o bin/goexplore.exe .
}

WIN_INSTALLER="$(ls -1 bin/*installer*.exe 2>/dev/null | head -1 || true)"
if [ -z "${WIN_INSTALLER}" ]; then
  WIN_INSTALLER="$(ls -1 bin/*.exe 2>/dev/null | head -1 || true)"
fi
if [ -n "${WIN_INSTALLER}" ]; then
  install -m 0755 "${WIN_INSTALLER}" "${OUTDIR}/goexplore-windows-installer.exe"
fi

echo "==> Building Debian package..."
bash scripts/ci-deb.sh amd64

echo "==> amd64 build complete:"
ls -lh "${OUTDIR}/goexplore-linux-amd64" \
  "${OUTDIR}/goexplore-windows-installer.exe" \
  "${OUTDIR}/goexplore_${VERSION}_amd64.deb"
