#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck source=ci-common.sh
source scripts/ci-common.sh

require_version

echo "==> Preparing frontend (bindings + production bundle)..."
apply_product_version
install_wails3_cli
ensure_build_assets
sync_app_icons

retry 3 go mod download

echo "==> Generating frontend bindings..."
retry 3 wails3 generate bindings -clean=true

echo "==> Building frontend..."
cd frontend
npm ci
npm run build
cd ..

echo "==> Prepare complete."
ls -la frontend/dist/index.html frontend/bindings/goexplore/index.js 2>/dev/null || true
