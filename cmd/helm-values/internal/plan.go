package internal

import (
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-values/internal/charts"
	"helmschema/cmd/helm-values/internal/config"
	"helmschema/cmd/helm-values/internal/docs"
	"helmschema/cmd/helm-values/internal/jsonschema"
	"os"

	"github.com/sirupsen/logrus"
)

func NewPlan(
	docsCfg *config.DocsConfig,
	schemaCfg *config.SchemaConfig,
	chart *charts.Chart,
) *Plan {
	return &Plan{
		chart:     chart,
		docsCfg:   docsCfg,
		schemaCfg: schemaCfg,
	}
}

type Plan struct {
	docsCfg   *config.DocsConfig
	schemaCfg *config.SchemaConfig
	chart     *charts.Chart
}

func (p *Plan) LogIntent(logger *logrus.Logger) {
	logger.Debugf("plan: %s: DryRun=%t", p.chart.Details.Name, p.DryRun())
	logger.Debugf("plan: %s: StrictComments=%t", p.chart.Details.Name, p.StrictComments())
	logger.Debugf("plan: %s: Stdout=%t", p.chart.Details.Name, p.StdOut())
	if p.docsCfg != nil {
		markup, err := p.DocsMarkup()
		logger.Debugf("plan: %s: UseDefault=%t", p.chart.Details.Name, p.DocsUseDefault())
		logger.Debugf("plan: %s: Markup=%s (error: %v)", p.chart.Details.Name, markup, err)
		logger.Debugf("plan: %s: Template=%s", p.chart.Details.Name, p.docsCfg.Template())
		logger.Debugf("plan: %s: Output=%v", p.chart.Details.Name, p.DocsOutputPath())
	}
	logger.Debugf("plan: %s: ChartRoot=%s", p.chart.Details.Name, p.chart.RootPath())
	logger.Debugf("plan: %s: ChartFile=%s", p.chart.Details.Name, p.chart.ChartFilePath())
	logger.Debugf("plan: %s: ChartValuesFile=%s", p.chart.Details.Name, p.chart.ValuesFilePath())
	logger.Debugf("plan: %s: ChartSchemaFile=%s", p.chart.Details.Name, p.chart.SchemaFilePath())
	logger.Debugf("plan: %s: ChartReadmeTemplate=%s", p.chart.Details.Name, p.DocsChartReadmeTemplate())
}

func (p *Plan) Chart() *charts.Chart {
	return p.chart
}

func (p *Plan) StdOut() bool {
	if p.docsCfg != nil {
		return p.docsCfg.StdOut()
	}
	if p.schemaCfg != nil {
		return p.schemaCfg.StdOut()
	}
	panic("no configs set")
}

func (p *Plan) StrictComments() bool {
	if p.docsCfg != nil {
		return p.docsCfg.Strict()
	}
	if p.schemaCfg != nil {
		return p.schemaCfg.Strict()
	}
	panic("no configs set")
}

func (p *Plan) DryRun() bool {
	if p.docsCfg != nil {
		return p.docsCfg.DryRun()
	}
	if p.schemaCfg != nil {
		return p.schemaCfg.DryRun()
	}
	panic("no configs set")
}

func (p *Plan) DocsChartReadmeTemplate() string {
	if _, err := os.Stat(p.chart.ReadmeMdTemplateFilePath()); err == nil {
		return p.chart.ReadmeMdTemplateFilePath()
	}
	if _, err := os.Stat(p.chart.ReadmeRstTemplateFilePath()); err == nil {
		return p.chart.ReadmeRstTemplateFilePath()
	}
	return ""
}

func (p *Plan) DocsMarkup() (docs.Markup, error) {
	markup, set, err := p.docsCfg.Markup()
	if err != nil {
		return "", err
	}
	if set {
		return markup, nil
	}

	if p.DocsUseDefault() {
		return docs.Markdown, nil
	}

	// If a template was specified, infer the markup type from that
	if tmpl := p.docsCfg.Template(); tmpl != "" {
		return docs.MarkupFromPath(tmpl)
	}

	// If there's a readme template in the chart, infer the markup type from that
	if tmpl := p.DocsChartReadmeTemplate(); tmpl != "" {
		return docs.MarkupFromPath(tmpl)
	}

	return "", fmt.Errorf("unable to infer markup type")
}

func (p *Plan) DocsUseDefault() bool {
	// If the user explicitly sets use-default, use that value
	if p.docsCfg.IsSet("use-default") {
		return p.docsCfg.GetBool("use-default")
	}

	// If a custom template file was set, use that
	if p.docsCfg.Template() != "" {
		return false
	}

	// If a custom template file is present, use that
	if p.DocsChartReadmeTemplate() != "" {
		return false
	}

	return true
}

func (p *Plan) DocsOutputPath() string {
	if p.docsCfg.IsSet("output") {
		return p.docsCfg.GetString("output")
	}

	docType, err := p.DocsMarkup()
	if err != nil {
		panic(err)
	}

	if docType == docs.Markdown {
		return p.chart.ReadmeMdFilePath()
	}
	if docType == docs.ReStructuredText {
		return p.chart.ReadmeRstFilePath()
	}

	panic("no readme template found")
}

func (p *Plan) WriteSchema(logger *logrus.Logger, schema *jsonschema.Schema) error {
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

func (p *Plan) WriteReadme(logger *logrus.Logger, s string) error {
	if !p.DryRun() {
		f, err := os.Create(p.DocsOutputPath())
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = f.Write([]byte(s)); err != nil {
			return err
		}
	}

	if p.StdOut() {
		fmt.Println(s)
	}

	return nil
}
