package docs

import (
	"helmschema/cmd/helm-values/internal/charts"
	"helmschema/cmd/helm-values/internal/jsonschema"
)

// type ValuesTable struct {
// 	Values []ValuesRow
// }

type ValuesRow struct {
	Key         string
	Type        string
	Default     string
	Description string
}

type RawContext struct {
	Chart  *charts.Chart
	Values *jsonschema.Schema
}

type TemplateContext struct {
	Raw         *RawContext
	ValuesTable []ValuesRow
}
