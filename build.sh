#!/usr/bin/env bash
# Builds md2pdf for Linux and Windows (amd64 each) into dist/.
#
#   ./build.sh                 # version from the VERSION file
#   ./build.sh 1.2.0           # version 1.2.0
#   VERSION=1.2.0 ./build.sh   # the same, via the environment
#
# The binaries are linked statically (CGO_ENABLED=0) and need no further
# libraries. At runtime md2pdf still needs an installed Chromium engine, which
# is not bundled.

set -euo pipefail

cd "$(dirname "$0")"

# The VERSION file at the repository root is the single source of truth; an
# argument or the environment overrides it for one-off builds.
FILE_VERSION="dev"
if [ -r VERSION ]; then
  FILE_VERSION="$(tr -d '[:space:]' < VERSION)"
fi
VERSION="${1:-${VERSION:-$FILE_VERSION}}"
OUTDIR="dist"
PKG="github.com/gstuebner/md2pdf"
LDFLAGS="-s -w -X ${PKG}/cmd.version=${VERSION}"

if ! command -v go >/dev/null 2>&1; then
  echo "build.sh: go is not installed or not in PATH." >&2
  exit 1
fi

mkdir -p "$OUTDIR"

build() {
  local goos="$1" goarch="$2" out="$3"
  printf '  %-16s ' "${goos}/${goarch}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "$LDFLAGS" -o "${OUTDIR}/${out}" .
  printf '%s (%s)\n' "${OUTDIR}/${out}" "$(du -h "${OUTDIR}/${out}" | cut -f1)"
}

echo "md2pdf ${VERSION} — $(go version | cut -d' ' -f3)"
echo

echo "Checking:"
printf '  %-16s ' "vet"
go vet ./...
echo "ok"
printf '  %-16s ' "test"
go test ./... >/dev/null
echo "ok"
echo

echo "Building:"
build linux   amd64 md2pdf
build windows amd64 md2pdf.exe
echo

echo "Done. Quick check of the Linux build:"
"${OUTDIR}/md2pdf" --version
