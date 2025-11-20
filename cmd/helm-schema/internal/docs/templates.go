package docs

import (
	"io"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
)

func NewTemplator(logger *logrus.Logger, chartName string, path string, extraPaths []string) (*Templator, error) {
	paths := append([]string{path}, extraPaths...)
	t, err := template.New(filepath.Base(path)).
		Funcs(sprig.FuncMap()).
		ParseFiles(paths...)
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
