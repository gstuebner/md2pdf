package model

import "html/template"

type Meta struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Kicker   string `yaml:"kicker"` // line above the title, defaults to "Documentation"
	Version  string `yaml:"version"`
	Author   string `yaml:"author"`
	Company  string `yaml:"company"`
	Date     string `yaml:"date"`   // display form, e.g. "2026-09-05"
	Logo     string `yaml:"logo"`   // path relative to the Markdown file
	Lang     string `yaml:"lang"`   // "de" or "en"; defaults to the detected locale
	Preset   string `yaml:"preset"` // style preset; --preset wins over this
}

type TOCEntry struct {
	Level int    // 1..6
	Text  string // plain text, no markup
	ID    string // anchor id
}

type Document struct {
	Meta       Meta
	TOC        []TOCEntry
	BodyHTML   template.HTML
	HasMermaid bool
}
