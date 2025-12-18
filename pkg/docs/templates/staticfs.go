package templates

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var TemplateFS embed.FS

func StaticTemplates() ([]string, error) {
	return fs.Glob(TemplateFS, "static/**/*.gotmpl")
}
