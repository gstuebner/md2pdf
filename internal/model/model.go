package model

import "html/template"

type Meta struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	Kicker   string `yaml:"kicker"` // Zeile über dem Titel, Default "Dokumentation"
	Version  string `yaml:"version"`
	Author   string `yaml:"author"`
	Company  string `yaml:"company"`
	Date     string `yaml:"date"` // Anzeigeform, z. B. "05.09.2026"
	Logo     string `yaml:"logo"` // Pfad relativ zur Markdown-Datei
	Lang     string `yaml:"lang"` // Default "de"
}

type TOCEntry struct {
	Level int    // 1..6
	Text  string // Klartext, ohne Markup
	ID    string // Anker-ID
}

type Document struct {
	Meta       Meta
	TOC        []TOCEntry
	BodyHTML   template.HTML
	HasMermaid bool
}
