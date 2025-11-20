package internal

import (
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-schema/internal/charts"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"os"

	"github.com/sirupsen/logrus"
)

func NewPlan(
	chart *charts.Chart,
	stdout bool,
	strictComments bool,
	dryRun bool,
) *Plan {
	return &Plan{
		chart:          chart,
		stdout:         stdout,
		strictComments: strictComments,
		dryRun:         dryRun,
	}
}

type Plan struct {
	strictComments bool
	stdout         bool
	dryRun         bool
	chart          *charts.Chart
}

func (p *Plan) LogIntent(logger *logrus.Logger) {
	logger.Debugf("%s: plan: DryRun=%t", p.chart.Details.Name, p.dryRun)
	logger.Debugf("%s: plan: StrictComments=%t", p.chart.Details.Name, p.strictComments)
	logger.Debugf("%s: plan: Stdout=%t", p.chart.Details.Name, p.stdout)
	logger.Debugf("%s: plan chart: root=%s", p.chart.Details.Name, p.chart.RootPath())
	logger.Debugf("%s: plan chart: ChartFile=%s", p.chart.Details.Name, p.chart.ChartFilePath())
	logger.Debugf("%s: plan chart: ValuesFile=%s", p.chart.Details.Name, p.chart.ValuesFilePath())
	logger.Debugf("%s: plan chart: SchemaFile=%s", p.chart.Details.Name, p.chart.SchemaFilePath())
	logger.Debugf("%s: plan chart: ReadmeMdTemplate=%s", p.chart.Details.Name, p.chart.ReadmeMdTemplateFilePath())
	logger.Debugf("%s: plan chart: ReadmeRstTemplate=%s", p.chart.Details.Name, p.chart.ReadmeRstTemplateFilePath())
}

func (p *Plan) Chart() *charts.Chart {
	return p.chart
}

func (p *Plan) StdOut() bool {
	return p.stdout
}

func (p *Plan) StrictComments() bool {
	return p.strictComments
}

func (p *Plan) DryRun() bool {
	return p.dryRun
}

func (p *Plan) ReadmeFilePath() string {
	if p.GenerateMarkdown() {
		return p.chart.ReadmeMdFilePath()
	}

	if p.GenerateRst() {
		return p.chart.ReadmeRstFilePath()
	}

	panic("no readme template found")
}

func (p *Plan) ReadmeTemplateFilePath() string {
	if p.GenerateMarkdown() {
		return p.chart.ReadmeMdTemplateFilePath()
	}

	if p.GenerateRst() {
		return p.chart.ReadmeRstTemplateFilePath()
	}

	panic("no readme template found")
}

func (p *Plan) GenerateMarkdown() bool {
	_, err := os.Stat(p.chart.ReadmeMdTemplateFilePath())
	return err == nil
}

func (p *Plan) GenerateRst() bool {
	_, err := os.Stat(p.chart.ReadmeRstTemplateFilePath())
	return err == nil
}

func (p *Plan) WriteSchema(logger *logrus.Logger, schema *jsonschema.Schema) error {
	s, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if p.stdout {
		fmt.Println(string(s))
	}

	if p.dryRun {
		return nil
	}

	logger.Infof("%s: schema: writing schema file", p.chart.Details.Name)
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

func (p *Plan) WriteReadme(logger *logrus.Logger, s string) error {
	if p.StdOut() {
		fmt.Println(s)
	}

	if p.DryRun() {
		return nil
	}

	logger.Debugf("%s: docs: opening readme file", p.Chart().Details.Name)
	f, err := os.Create(p.ReadmeFilePath())
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write([]byte(s)); err != nil {
		return err
	}

	return nil
}
