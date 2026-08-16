// Package templates provides the template building blocks used to render chart
// documentation, including the built-in static templates, template functions, and
// markup-type handling.
package templates

import (
	"errors"
	"strings"
)

// Markup is an enum describing the supported documentation markup types.
type Markup string

const (
	// Markdown is the Markdown documentation markup type.
	Markdown Markup = "markdown"
	// ReStructuredText is the reStructuredText documentation markup type.
	ReStructuredText Markup = "restructuredtext"
)

// MarkupFromString parses a Markup from its canonical or shorthand string
// representation (e.g. "markdown"/"md", "restructuredtext"/"rst").
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

// MarkupFromPath infers a Markup from a template file path's extension.
func MarkupFromPath(path string) (Markup, error) {
	if strings.Contains(path, ".md.tmpl") || strings.Contains(path, ".md.gotmpl") {
		return Markdown, nil
	}

	if strings.Contains(path, ".rst.tmpl") || strings.Contains(path, ".rst.gotmpl") {
		return ReStructuredText, nil
	}

	return "", errors.New("unable to infer markup type")
}
