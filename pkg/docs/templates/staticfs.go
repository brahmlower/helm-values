package templates

import (
	"embed"
	"fmt"
	"io/fs"
)

// TemplateFS embeds the built-in static templates shipped with the binary.
//
//go:embed all:static
var TemplateFS embed.FS

// StaticTemplates returns the paths of the built-in static templates embedded in
// TemplateFS.
func StaticTemplates() ([]string, error) {
	matches, err := fs.Glob(TemplateFS, "static/**/*.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("globbing static templates: %w", err)
	}

	return matches, nil
}
