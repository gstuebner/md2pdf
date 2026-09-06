# Design layer

This is where the reference for the look of the generated PDFs lives. The
actual assets that get embedded into the binary live under
`internal/assets/`.

## Files

| File                    | Purpose                                                                |
| ----------------------- | --------------------------------------------------------------------- |
| `preview.html`          | Static design reference with all supported elements, with a preset switcher |
| `branding-example.css`  | Example of how to recolor the theme via `--css`                       |

`preview.html` pulls in the theme files from `internal/assets/theme/` via
`<link>`. So there's no second copy of the CSS — what you see here is
exactly what the binary ships later on. The switcher in the top right corner
swaps the preset stylesheet, so every preset can be reviewed in the same
document.

## Viewing

```fish
xdg-open design/preview.html
```

## Checking as a PDF

```fish
chromium --headless=new --disable-gpu --no-pdf-header-footer \
  --print-to-pdf=/tmp/preview.pdf design/preview.html
```

Header and footer are missing here because the CLI route doesn't know about
templates. Only `md2pdf` itself, via the DevTools protocol, produces the
full result including the running header and page numbers.

## Elements in the preview

Cover page, table of contents, H2–H4 headings with numbering, body text,
bullet lists (including nested ones), numbered lists, task lists,
definition lists, blockquote, all five callout types, tables, inline code,
code blocks (bash, yaml, go, shell, diff, text), a figure with caption and
numbering, two Mermaid diagrams, footnotes.

## Stylesheet layout

The theme is split in two. `internal/assets/theme/base.css` carries the
structure every preset shares — layout, page handling, counters, and the
mechanics behind cover, contents, callouts, tables, code blocks and figures.
Colours, fonts, sizes and decoration come from
`internal/assets/theme/presets/<name>.css`, which is appended after it, so a
preset can also override single rules. `code.css` (Chroma colours) and the
`--css` files come last.

A new preset is one file under `theme/presets/` plus one entry in
`config.Presets` (`internal/config/preset.go`), which also carries the
structural defaults — cover page, contents, numbering, margins.

## Swapping fonts

The fonts live as woff2 under `internal/assets/fonts/` and get written by
`scripts/genfonts.sh` as data URIs into `internal/assets/theme/fonts.css`.
After swapping a woff2 file:

```fish
./scripts/genfonts.sh
```

All bundled fonts — Inter, Source Serif 4 and JetBrains Mono — are licensed
under the SIL Open Font License 1.1; the license texts sit right next to
them.
