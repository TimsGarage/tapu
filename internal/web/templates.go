package web

import (
	"embed"
	"html/template"
)

// templateFiles contains the hosted IAM pages.
//
//go:embed templates/*.html
var templateFiles embed.FS

func parseTemplates() (*template.Template, error) {
	return template.ParseFS(templateFiles, "templates/*.html")
}
