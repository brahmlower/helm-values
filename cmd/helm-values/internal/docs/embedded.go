package docs

import (
	"embed"
	"io/fs"
)

//go:embed all:templates
var TemplateFS embed.FS

func StaticTemplates() ([]string, error) {
	return fs.Glob(TemplateFS, "templates/**/*.gotmpl")
}
