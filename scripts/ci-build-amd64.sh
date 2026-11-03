#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${VERSION:-}" ]; then
  echo "VERSION must be set" >&2
  exit 1
fi

OUTDIR="${OUTDIR:-dist}"
mkdir -p "${OUTDIR}" bin

export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"

retry() {
  local attempts=$1
  shift
  local n=1
  until "$@"; do
    if [ "$n" -ge "$attempts" ]; then
      return 1
    fi
    n=$((n + 1))
    echo "Command failed (attempt ${n}/${attempts}), retrying..." >&2
    sleep 3
  done
}

LINUX_PACKAGES="gcc pkg-config libgtk-4-dev libwebkitgtk-6.0-dev mingw-w64 nsis"

echo "==> Installing build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq ${LINUX_PACKAGES}

echo "==> Installing Wails v3 CLI..."
WAILS_VERSION="$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)"
retry 3 env CGO_ENABLED=0 go install "github.com/wailsapp/wails/v3/cmd/wails3@${WAILS_VERSION}"
export PATH="$(go env GOPATH)/bin:${PATH}"

echo "==> Setting product version..."
sed -i "s/\"productVersion\": \".*\"/\"productVersion\": \"${VERSION}\"/" wails.json
if [ -f build/config.yml ]; then
  sed -i "s/^  version: \".*\"/  version: \"${VERSION}\"/" build/config.yml
fi

if [ ! -f build/Taskfile.yml ]; then
  echo "==> Generating Wails v3 build assets..."
  wails3 generate build-assets -dir build -name GoExplore -binaryname goexplore \
    -productname GoExplore -productcompany "JDB-NET Limited" -productversion "${VERSION}"
fi
echo "==> Updating Wails v3 build assets..."
wails3 update build-assets -name GoExplore -binaryname goexplore -config ./wails.json -dir build

echo "==> Syncing app icon (generate build-assets ships the default Wails artwork)..."
cp frontend/src/assets/appicon.png build/appicon.png
(
  cd build
  wails3 generate icons -input appicon.png \
    -windowsfilename windows/icon.ico \
    -macfilename darwin/icons.icns
)

echo "==> Downloading Go modules..."
retry 3 go mod download

echo "==> Generating frontend bindings..."
retry 3 wails3 generate bindings -clean=true

echo "==> Building frontend..."
cd frontend
npm ci
npm run build
cd ..

LDFLAGS="-w -s -X main.Version=v${VERSION}"
export CGO_CFLAGS="-Wno-deprecated-declarations${CGO_CFLAGS:+ ${CGO_CFLAGS}}"

echo "==> Building Linux amd64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags production -trimpath -ldflags "${LDFLAGS}" -o "${OUTDIR}/goexplore-linux-amd64" .

WIN_BUILD_FLAGS="-tags production -trimpath -buildvcs=false -ldflags \"-w -s -H windowsgui -X main.Version=v${VERSION}\""

echo "==> Building Windows amd64 installer..."
retry 2 wails3 task windows:package GOOS=windows ARCH=amd64 BUILD_FLAGS="${WIN_BUILD_FLAGS}"

WIN_INSTALLER="build/windows/nsis/goexplore-installer.exe"
if [ ! -f "${WIN_INSTALLER}" ]; then
  echo "Windows installer not found at ${WIN_INSTALLER}" >&2
  exit 1
fi
install -m 0755 "${WIN_INSTALLER}" "${OUTDIR}/goexplore-windows-installer.exe"

echo "==> Building Debian package..."
bash scripts/ci-deb.sh amd64

echo "==> amd64 build complete:"
ls -lh "${OUTDIR}/goexplore-linux-amd64" \
  "${OUTDIR}/goexplore-windows-installer.exe" \
  "${OUTDIR}/goexplore_${VERSION}_amd64.deb"
