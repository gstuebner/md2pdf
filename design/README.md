# Designschicht

Hier liegt die Referenz fuer das Aussehen der erzeugten PDFs. Die eigentlichen
Assets, die ins Binary eingebettet werden, liegen unter `internal/assets/`.

## Dateien

| Datei                   | Zweck                                                                 |
| ----------------------- | --------------------------------------------------------------------- |
| `preview.html`          | Statische Designreferenz mit allen unterstuetzten Elementen            |
| `branding-example.css`  | Beispiel, wie sich das Theme ueber `--css` umfaerben laesst            |

`preview.html` bindet die Theme-Dateien aus `internal/assets/theme/` per
`<link>` ein. Es gibt also keine zweite Kopie des CSS — was hier zu sehen ist,
ist genau das, was das Binary spaeter ausliefert.

## Ansehen

```fish
xdg-open design/preview.html
```

## Als PDF pruefen

```fish
chromium --headless=new --disable-gpu --no-pdf-header-footer \
  --print-to-pdf=/tmp/preview.pdf design/preview.html
```

Kopf- und Fusszeile fehlen dabei, weil der CLI-Weg keine Templates kennt. Das
vollstaendige Ergebnis inklusive laufender Kopfzeile und Seitenzahlen erzeugt
erst `md2pdf` selbst ueber das DevTools-Protokoll.

## Elemente in der Vorschau

Deckblatt, Inhaltsverzeichnis, Ueberschriften H2–H4 mit Nummerierung,
Fliesstext, Aufzaehlungen (auch verschachtelt), nummerierte Listen,
Aufgabenlisten, Definitionslisten, Zitat, alle fuenf Callout-Arten,
Tabellen, Inline-Code, Codebloecke (bash, yaml, go, shell, diff, text),
Abbildung mit Bildunterschrift und Nummerierung, zwei Mermaid-Diagramme,
Fussnoten.

## Schriften austauschen

Die Fonts liegen als woff2 unter `internal/assets/fonts/` und werden von
`scripts/genfonts.sh` als Data-URIs nach `internal/assets/theme/fonts.css`
geschrieben. Nach dem Austausch einer woff2-Datei:

```fish
./scripts/genfonts.sh
```

Beide mitgelieferten Schriften stehen unter der SIL Open Font License 1.1,
die Lizenztexte liegen daneben.
