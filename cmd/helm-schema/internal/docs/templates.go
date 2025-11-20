package docs

import (
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
)

func NewTemplator(logger *logrus.Logger, fsys fs.FS, chartName string, path string, extraPaths []string) (*Templator, error) {
	paths := append([]string{path}, extraPaths...)

	// A ridiculously stupid hack to get file lookups working
	// through the Root.FS because full absolute paths don't seem
	// to work when the root directory itself is mounted.
	for i, p := range paths {
		paths[i] = strings.TrimPrefix(p, "/")
	}

	t, err := template.New(filepath.Base(path)).
		Funcs(sprig.FuncMap()).
		ParseFS(fsys, paths...)
	if err != nil {
		return nil, err
	}

	templator := &Templator{
		target: t,
	}

	return templator, nil
}

type Templator struct {
	target *template.Template
}

func (t *Templator) Render(w io.Writer, data TemplateContext) error {
	return t.target.Execute(w, data)
}
