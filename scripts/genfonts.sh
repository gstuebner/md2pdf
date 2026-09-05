#!/usr/bin/env bash
# Generiert internal/assets/theme/fonts.css aus den woff2-Dateien in internal/assets/fonts/.
set -euo pipefail
cd "$(dirname "$0")/.."
F=internal/assets/fonts
OUT=internal/assets/theme/fonts.css

LATIN='U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,U+0304,U+0308,U+0329,U+2000-206F,U+20AC,U+2122,U+2191,U+2193,U+2212,U+2215,U+FEFF,U+FFFD'
LATINEXT='U+0100-02AF,U+0304,U+0308,U+0329,U+1E00-1E9F,U+1EF2-1EFF,U+2020,U+20A0-20AB,U+20AD-20C0,U+2113,U+2C60-2C7F,U+A720-A7FF'

face() { # family style range file
  printf '@font-face {\n  font-family: "%s";\n  font-style: %s;\n  font-weight: 100 900;\n  font-display: block;\n  unicode-range: %s;\n  src: url(data:font/woff2;base64,%s) format("woff2");\n}\n\n' \
    "$1" "$2" "$3" "$(base64 -w0 "$F/$4")"
}

{
  printf '/* Erzeugt von scripts/genfonts.sh — nicht von Hand bearbeiten.\n'
  printf '   Inter und JetBrains Mono, SIL Open Font License 1.1.\n'
  printf '   Lizenztexte: internal/assets/fonts/OFL-*.txt */\n\n'
  face "Inter"          normal "$LATIN"    inter-latin-wght-normal.woff2
  face "Inter"          normal "$LATINEXT" inter-latin-ext-wght-normal.woff2
  face "Inter"          italic "$LATIN"    inter-latin-wght-italic.woff2
  face "JetBrains Mono" normal "$LATIN"    jetbrains-mono-latin-wght-normal.woff2
} > "$OUT"

printf 'fonts.css: %s bytes\n' "$(stat -c%s "$OUT")"
