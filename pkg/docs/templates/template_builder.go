package templates

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// DefaultMarkdownTemplate is the name of the built-in template used to render
// Markdown documentation when no custom template is configured.
const DefaultMarkdownTemplate = "default.md.gotmpl"

// DefaultReStructuredTextTemplate is the name of the built-in template used to render
// reStructuredText documentation when no custom template is configured.
const DefaultReStructuredTextTemplate = "default.rst.gotmpl"

// TemplateBuilder assembles a text/template.Template from a set of static, extra, and
// optionally custom template paths, ready to render a chart's documentation.
type TemplateBuilder struct {
	customTemplate string
	extraPaths     []string
	useDefault     bool
	markup         Markup
}

// NewTemplateBuilder creates a TemplateBuilder configured with the given options.
func NewTemplateBuilder(opts ...BuilderOpt) *TemplateBuilder {
	t := &TemplateBuilder{
		customTemplate: "",
		extraPaths:     nil,
		useDefault:     false,
		markup:         "",
	}
	for _, s := range opts {
		s(t)
	}

	return t
}

// TemplateName returns the name of the root template to execute: the built-in default
// for the configured markup type when useDefault is set, otherwise the base name of the
// configured custom template.
func (b *TemplateBuilder) TemplateName() string {
	if b.useDefault && b.markup == Markdown {
		return DefaultMarkdownTemplate
	}

	if b.useDefault && b.markup == ReStructuredText {
		return DefaultReStructuredTextTemplate
	}

	return filepath.Base(b.customTemplate)
}

// TemplatePaths returns the full set of template paths to parse: the builder's extra
// paths, plus the custom template path when useDefault is not set.
func (b *TemplateBuilder) TemplatePaths() []string {
	paths := []string{}

	paths = append(paths, b.extraPaths...)
	if !b.useDefault {
		paths = append(paths, b.customTemplate)
	}

	return paths
}

// Build parses the builder's template paths from fsys, registering the builder's
// template functions, and returns the resulting root template.
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

	tmpl, err := template.New(b.TemplateName()).
		Funcs(funcMap).
		ParseFS(fsys, paths...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates %v: %w", paths, err)
	}

	return tmpl, nil
}

// WithCustomTemplate sets the builder's custom template path, disabling use of the
// built-in default template, and infers the markup type from the path when possible.
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

// WithExtraPaths sets the builder's extra template paths, which are always parsed
// alongside the default or custom template.
func WithExtraPaths(paths []string) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.extraPaths = paths
	}
}

// WithUseDefault sets whether the builder should render the built-in default template
// instead of a custom template.
func WithUseDefault(useDefault bool) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.useDefault = useDefault
	}
}

// WithMarkup sets the markup type the builder should render, used to select the
// appropriate built-in default template.
func WithMarkup(markup Markup) BuilderOpt {
	return func(t *TemplateBuilder) {
		t.markup = markup
	}
}

// BuilderOpt configures a TemplateBuilder, applied by NewTemplateBuilder.
type BuilderOpt = func(*TemplateBuilder)
