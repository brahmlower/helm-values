package templates

import (
	"helmschema/internal/charts"
	"helmschema/pkg/schema"
)

type ValuesRow struct {
	Key         string
	Type        string
	Default     string
	Description string
}

type RawContext struct {
	Chart  *charts.Chart
	Values *schema.Schema
}

type TemplateContext struct {
	Raw         *RawContext
	ValuesTable []ValuesRow
}
