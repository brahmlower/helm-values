package templates

import (
	"helmvalues/pkg"
	"helmvalues/pkg/charts"
)

// ValuesRow is one rendered row of a chart's values table.
type ValuesRow struct {
	Key         string
	Type        string
	Default     string
	Description string
}

// RawContext exposes the underlying chart and its parsed values schema to
// templates.
type RawContext struct {
	Chart  *charts.Chart
	Values *pkg.JsonSchema
}

// TemplateContext is the data made available to a documentation template.
type TemplateContext struct {
	Raw         *RawContext
	ValuesTable []ValuesRow
}
