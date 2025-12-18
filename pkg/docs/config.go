package docs

import (
	"fmt"
	"strings"

	"helmschema/pkg/docs/templates"

	"github.com/samber/mo"
	"github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel       logrus.Level
	StdOut         bool
	Strict         bool
	DryRun         bool
	UseDefault     mo.Option[bool]
	Output         mo.Option[string]
	Template       string
	ExtraTemplates []string
	Markup         mo.Option[templates.Markup]
	Order          ValuesOrder
}

type ValuesOrder string

const (
	ValuesOrderAlphabetical ValuesOrder = "alphabetical"
	ValuesOrderPreserve     ValuesOrder = "preserve"
)

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
