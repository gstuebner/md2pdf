// Command md2pdf converts a Markdown file into a print-ready PDF.
package main

import (
	"os"

	"github.com/gstuebner/md2pdf/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
