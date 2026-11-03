#!/usr/bin/env bash
# Shared helpers for release CI scripts. Source from repo root after cd to repo.

set -euo pipefail

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

require_version() {
  if [ -z "${VERSION:-}" ]; then
    echo "VERSION must be set" >&2
    exit 1
  fi
}

apply_product_version() {
  sed -i "s/\"productVersion\": \".*\"/\"productVersion\": \"${VERSION}\"/" wails.json
  if [ -f build/config.yml ]; then
    sed -i "s/^  version: \".*\"/  version: \"${VERSION}\"/" build/config.yml
  fi
}

install_wails3_cli() {
  local wails_version
  wails_version="$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)"
  retry 3 env CGO_ENABLED=0 go install "github.com/wailsapp/wails/v3/cmd/wails3@${wails_version}"
  export PATH="$(go env GOPATH)/bin:${PATH}"
}

ensure_build_assets() {
  if [ ! -f build/Taskfile.yml ]; then
    echo "==> Generating Wails v3 build assets..."
    wails3 generate build-assets -dir build -name GoExplore -binaryname goexplore \
      -productname GoExplore -productcompany "JDB-NET Limited" -productversion "${VERSION}"
  fi
  echo "==> Updating Wails v3 build assets..."
  wails3 update build-assets -name GoExplore -binaryname goexplore -config ./wails.json -dir build
}

sync_app_icons() {
  echo "==> Syncing app icon..."
  cp frontend/src/assets/appicon.png build/appicon.png
  (
    cd build
    wails3 generate icons -input appicon.png \
      -windowsfilename windows/icon.ico \
      -macfilename darwin/icons.icns
  )
}
