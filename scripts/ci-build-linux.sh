#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck source=ci-common.sh
source scripts/ci-common.sh

require_version

OUTDIR="${OUTDIR:-dist}"
mkdir -p "${OUTDIR}"

if [ ! -f frontend/dist/index.html ]; then
  echo "Missing frontend/dist; run prepare or download the frontend-build artifact." >&2
  exit 1
fi

LINUX_PACKAGES="gcc pkg-config libgtk-4-dev libwebkitgtk-6.0-dev"

echo "==> Installing Linux build dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq ${LINUX_PACKAGES}

apply_product_version
sync_app_icons
retry 3 go mod download

LDFLAGS="-w -s -X main.Version=v${VERSION}"
export CGO_CFLAGS="-Wno-deprecated-declarations${CGO_CFLAGS:+ ${CGO_CFLAGS}}"

echo "==> Building Linux amd64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags production -trimpath -ldflags "${LDFLAGS}" -o "${OUTDIR}/goexplore-linux-amd64" .

echo "==> Building Debian package..."
bash scripts/ci-deb.sh amd64

echo "==> Portable Linux launcher (KDE taskbar icon when not using menu install)..."
cp build/appicon.png "${OUTDIR}/goexplore.png"
portable_desktop="${OUTDIR}/GoExplore.desktop"
sed \
  -e 's|@EXEC@|./goexplore-linux-amd64|g' \
  -e 's|@ICON@|./goexplore.png|g' \
  packaging/linux/GoExplore-portable.desktop >"${portable_desktop}.tmp"
mv "${portable_desktop}.tmp" "${portable_desktop}"

ls -lh "${OUTDIR}/goexplore-linux-amd64" "${OUTDIR}"/goexplore_"${VERSION}"_*.deb \
  "${OUTDIR}/goexplore.png" "${OUTDIR}/GoExplore.desktop"
