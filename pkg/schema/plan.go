package schema

import (
	"encoding/json"
	"fmt"
	"helmschema/internal/charts"
	"os"

	"github.com/sirupsen/logrus"
)

func NewPlan(cfg *Config, chart *charts.Chart) *Plan {
	return &Plan{
		chart: chart,
		cfg:   cfg,
	}
}

type Plan struct {
	cfg   *Config
	chart *charts.Chart
}

func (p *Plan) LogCommonDetails(logger *logrus.Logger) {
	// common configs
	logger.Debugf("plan: %s: DryRun=%t", p.chart.Details.Name, p.DryRun())
	logger.Debugf("plan: %s: StrictComments=%t", p.chart.Details.Name, p.StrictComments())
	logger.Debugf("plan: %s: Stdout=%t", p.chart.Details.Name, p.StdOut())
}

func (p *Plan) LogChartDetails(logger *logrus.Logger) {
	// chart configs
	logger.Debugf("plan: %s: ChartRoot=%s", p.chart.Details.Name, p.chart.RootPath())
	logger.Debugf("plan: %s: ChartFile=%s", p.chart.Details.Name, p.chart.ChartFilePath())
	logger.Debugf("plan: %s: ChartValuesFile=%s", p.chart.Details.Name, p.chart.ValuesFilePath())
	logger.Debugf("plan: %s: ChartSchemaFile=%s", p.chart.Details.Name, p.chart.SchemaFilePath())
	// logger.Debugf("plan: %s: ChartReadmeTemplate=%s", p.chart.Details.Name, p.DocsChartReadmeTemplate())
}

func (p *Plan) LogSchemaDetails(logger *logrus.Logger) {
	logger.Debugf("plan: %s: WriteModeline=%t", p.chart.Details.Name, p.cfg.WriteModeline)
}

func (p *Plan) Chart() *charts.Chart {
	return p.chart
}

func (p *Plan) StdOut() bool {
	return p.cfg.StdOut
}

func (p *Plan) StrictComments() bool {
	return p.cfg.Strict
}

func (p *Plan) DryRun() bool {
	return p.cfg.DryRun
}

func (p *Plan) WriteSchema(logger *logrus.Logger, schema *Schema) error {
	s, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if p.StdOut() {
		fmt.Println(string(s))
	}

	if p.DryRun() {
		return nil
	}

	f, err := os.Create(p.chart.SchemaFilePath())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(string(s))
	if err != nil {
		return err
	}

	return nil
}
