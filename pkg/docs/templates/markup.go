package templates

import (
	"errors"
	"strings"
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
