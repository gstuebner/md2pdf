#!/usr/bin/env bash
# Baut md2pdf für Linux und Windows (jeweils amd64) nach dist/.
#
#   ./build.sh                 # Version "dev"
#   ./build.sh 1.2.0           # Version 1.2.0
#   VERSION=1.2.0 ./build.sh   # dasselbe über die Umgebung
#
# Es wird statisch gelinkt (CGO_ENABLED=0), das Ergebnis läuft also ohne
# weitere Bibliotheken. Zur Laufzeit braucht md2pdf trotzdem eine installierte
# Chromium-Engine — die wird nicht mitgeliefert.

set -euo pipefail

cd "$(dirname "$0")"

VERSION="${1:-${VERSION:-dev}}"
OUTDIR="dist"
PKG="github.com/gstuebner/md2pdf"
LDFLAGS="-s -w -X ${PKG}/cmd.version=${VERSION}"

if ! command -v go >/dev/null 2>&1; then
  echo "build.sh: go ist nicht installiert oder nicht im PATH." >&2
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

echo "Prüfen:"
printf '  %-16s ' "vet"
go vet ./...
echo "ok"
printf '  %-16s ' "test"
go test ./... >/dev/null
echo "ok"
echo

echo "Bauen:"
build linux   amd64 md2pdf
build windows amd64 md2pdf.exe
echo

echo "Fertig. Schnelltest der Linux-Fassung:"
"${OUTDIR}/md2pdf" --version
