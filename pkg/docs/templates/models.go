package templates

import (
	"helmvalues/pkg"
	"helmvalues/pkg/charts"
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
