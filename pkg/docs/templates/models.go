package templates

import (
	"helmschema/internal/charts"
	"helmschema/pkg"
)

type ValuesRow struct {
	Key         string
	Type        string
	Default     string
	Description string
}

type RawContext struct {
	Chart  *charts.Chart
	Values *pkg.JsonSchema
}

type TemplateContext struct {
	Raw         *RawContext
	ValuesTable []ValuesRow
}
