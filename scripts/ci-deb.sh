#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${VERSION:-}" ]; then
  echo "VERSION must be set" >&2
  exit 1
fi

OUTDIR="${OUTDIR:-dist}"
mkdir -p "${OUTDIR}"

if ! command -v nfpm >/dev/null 2>&1; then
  go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.47.0
  export PATH="$(go env GOPATH)/bin:${PATH}"
fi

if [ "$#" -gt 0 ]; then
  arches=("$@")
else
  arches=(amd64)
fi

packaged=0
for arch in "${arches[@]}"; do
  binary="${OUTDIR}/goexplore-linux-${arch}"
  if [ ! -f "${binary}" ]; then
    echo "Missing ${binary}" >&2
    exit 1
  fi

  mkdir -p dist/nfpm-staging
  cp "${binary}" dist/nfpm-staging/goexplore

  NFPM_VERSION="${VERSION}" NFPM_ARCH="${arch}" nfpm package \
    -f nfpm/goexplore.yaml \
    -t "${OUTDIR}/goexplore_${VERSION}_${arch}.deb"

  rm -rf dist/nfpm-staging
  packaged=$((packaged + 1))
done

if [ "${packaged}" -eq 0 ]; then
  echo "No Debian packages were produced" >&2
  exit 1
fi

echo "==> Debian packages:"
ls -lh "${OUTDIR}"/goexplore_"${VERSION}"_*.deb
