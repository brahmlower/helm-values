package docs

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// enum describing template types
type Markup string

const (
	Markdown         Markup = "markdown"
	ReStructuredText Markup = "restructuredtext"
)

func MarkupFromString(s string) (Markup, error) {
	switch s {
	case "markdown", "md":
		return Markdown, nil
	case "restructuredtext", "rst":
		return ReStructuredText, nil
	default:
		return "", errors.New("invalid markup type")
	}
}

func MarkupFromPath(path string) (Markup, error) {
	if strings.Contains(path, ".md.tmpl") || strings.Contains(path, ".md.gotmpl") {
		return Markdown, nil
	}
	if strings.Contains(path, ".rst.tmpl") || strings.Contains(path, ".rst.gotmpl") {
		return ReStructuredText, nil
	}
	return "", errors.New("unable to infer markup type")
}

const DefaultMarkdownTemplate = "default.md.gotmpl"
const DefaultReStructuredTextTemplate = "default.rst.gotmpl"

type TemplateBuilder struct {
	customTemplate string
	extraPaths     []string
	useDefault     bool
	markup         Markup
}

func (b *TemplateBuilder) TemplateName() string {
	if b.useDefault && b.markup == Markdown {
		return DefaultMarkdownTemplate
	}
	if b.useDefault && b.markup == ReStructuredText {
		return DefaultReStructuredTextTemplate
	}
	return filepath.Base(b.customTemplate)
}

func (b *TemplateBuilder) TemplatePaths() []string {
	paths := []string{}
	paths = append(paths, b.extraPaths...)
	if !b.useDefault {
		paths = append(paths, b.customTemplate)
	}
	return paths
}

func (b *TemplateBuilder) Build(fsys fs.FS) (*template.Template, error) {
	paths := b.TemplatePaths()

	// A ridiculously stupid hack to get file lookups working
	// through the Root.FS because full absolute paths don't seem
	// to work when the root directory itself is mounted.
	for i, p := range paths {
		paths[i] = strings.TrimPrefix(p, "/")
	}

	funcMap := sprig.FuncMap()
	funcMap["lpad"] = lpad
	funcMap["rpad"] = rpad
	funcMap["maxLen"] = maxLen
	funcMap["rowSelect"] = rowSelect
	funcMap["mdRow"] = mdRow
	funcMap["mdMultiline"] = mdMultiline

	return template.New(b.TemplateName()).
		Funcs(funcMap).
		ParseFS(fsys, paths...)
}

func WithCustomTemplate(template string) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.customTemplate = template
		t.useDefault = false

		// Ignore errors here because it's just best effort
		if markup, err := MarkupFromPath(t.customTemplate); err == nil {
			t.markup = markup
		}
	}
}

func WithExtraPaths(paths []string) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.extraPaths = paths
	}
}

func WithUseDefault(useDefault bool) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.useDefault = useDefault
	}
}

func WithMarkup(markup Markup) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.markup = markup
	}
}

type BuilderOpt = func(*TemplateBuilder)

func NewTemplateBuilder(opts ...BuilderOpt) *TemplateBuilder {
	t := &TemplateBuilder{}
	for _, s := range opts {
		s(t)
	}
	return t
}
