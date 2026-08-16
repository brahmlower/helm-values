// Package docs generates a chart's values documentation (e.g. README.md)
// from its values schema.
package docs

import (
	"fmt"
	"strings"

	"helmvalues/pkg/docs/templates"

	"github.com/samber/mo"
	"github.com/sirupsen/logrus"
)

// Config controls how a chart's values documentation is generated.
type Config struct {
	LogLevel       logrus.Level
	StdOut         bool
	Strict         bool
	DryRun         bool
	GitAdd         bool
	UseDefault     mo.Option[bool]
	Output         mo.Option[string]
	Template       string
	ExtraTemplates []string
	Markup         mo.Option[templates.Markup]
	Order          ValuesOrder
}

// ValuesOrder controls the order in which values rows are rendered.
type ValuesOrder string

const (
	// ValuesOrderAlphabetical sorts values rows alphabetically by key.
	ValuesOrderAlphabetical ValuesOrder = "alphabetical"
	// ValuesOrderPreserve keeps values rows in their source file order.
	ValuesOrderPreserve ValuesOrder = "preserve"
)

// NewValuesOrder parses orderStr into a ValuesOrder.
func NewValuesOrder(orderStr string) (ValuesOrder, error) {
	switch strings.ToLower(orderStr) {
	case "alphabetical":
		return ValuesOrderAlphabetical, nil
	case "preserve":
		return ValuesOrderPreserve, nil
	default:
		return "", fmt.Errorf("invalid values order: %s", orderStr)
	}
}
