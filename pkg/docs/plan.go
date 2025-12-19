package docs

import (
	"errors"
	"fmt"
	"helmvalues/internal/charts"
	"helmvalues/pkg/docs/templates"
	"helmvalues/pkg/schema"
	"os"

	"github.com/sirupsen/logrus"
)

func NewPlan(cfg *Config, chart *charts.Chart) *Plan {
	schemaCfg := &schema.Config{
		StdOut:        cfg.StdOut,
		Strict:        cfg.Strict,
		DryRun:        cfg.DryRun,
		WriteModeline: false,
		LogLevel:      cfg.LogLevel,
	}
	schemaPlan := schema.NewPlan(schemaCfg, chart)

	return &Plan{
		chart:      chart,
		cfg:        cfg,
		schemaPlan: schemaPlan,
	}
}

type Plan struct {
	cfg        *Config
	chart      *charts.Chart
	schemaPlan *schema.Plan
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
	logger.Debugf("plan: %s: ChartReadmeTemplate=%s", p.chart.Details.Name, p.DocsChartReadmeTemplate())
}

func (p *Plan) LogDocDetails(logger *logrus.Logger) {
	logger.Debugf("plan: %s: UseDefault=%t", p.chart.Details.Name, p.DocsUseDefault())
	template, builtin, err := p.DocsTargetTemplate()
	logger.Debugf("plan: %s: Template=%s (default: %t, error: %v)", p.chart.Details.Name, template, builtin, err)
	markup, err := p.DocsMarkup()
	logger.Debugf("plan: %s: Markup=%s (error: %v)", p.chart.Details.Name, markup, err)
	outputPath, err := p.DocsOutputPath()
	logger.Debugf("plan: %s: Output=%s (error: %v)", p.chart.Details.Name, outputPath, err)
	logger.Debugf("plan: %s: ValuesOrder=%s (error: %v)", p.chart.Details.Name, p.cfg.Order, err)
}

func (p *Plan) LogSchemaDetails(logger *logrus.Logger) {
	p.schemaPlan.LogSchemaDetails(logger)
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

func (p *Plan) DocsTargetTemplate() (string, bool, error) {
	if p.cfg.Template != "" {
		return p.cfg.Template, false, nil
	}

	if tmpl := p.DocsChartReadmeTemplate(); tmpl != "" {
		return tmpl, false, nil
	}

	if p.DocsUseDefault() {
		return "", true, nil
	}

	return "", false, errors.New("no target template found")
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

func (p *Plan) DocsMarkup() (templates.Markup, error) {
	if value, ok := p.cfg.Markup.Get(); ok {
		return value, nil
	}

	if p.DocsUseDefault() {
		return templates.Markdown, nil
	}

	// If a template was specified, infer the markup type from that
	if tmpl := p.cfg.Template; tmpl != "" {
		return templates.MarkupFromPath(tmpl)
	}

	// If there's a readme template in the chart, infer the markup type from that
	if tmpl := p.DocsChartReadmeTemplate(); tmpl != "" {
		return templates.MarkupFromPath(tmpl)
	}

	return "", errors.New("unable to infer markup type")
}

func (p *Plan) DocsUseDefault() bool {
	// If the user explicitly sets use-default, use that value
	if useDefault, ok := p.cfg.UseDefault.Get(); ok {
		return useDefault
	}

	// If a custom template file was set, use that
	if p.cfg.Template != "" {
		return false
	}

	// If a custom template file is present, use that
	if p.DocsChartReadmeTemplate() != "" {
		return false
	}

	return true
}

func (p *Plan) DocsOutputPath() (string, error) {
	if output, ok := p.cfg.Output.Get(); ok {
		return output, nil
	}

	docType, err := p.DocsMarkup()
	if err != nil {
		return "", err
	}

	if docType == templates.Markdown {
		return p.chart.ReadmeMdFilePath(), nil
	}
	if docType == templates.ReStructuredText {
		return p.chart.ReadmeRstFilePath(), nil
	}

	panic("invalid markup type")
}

func (p *Plan) SchemaPlan() *schema.Plan {
	return p.schemaPlan
}

func (p *Plan) WriteReadme(logger *logrus.Logger, s string) error {
	if !p.DryRun() {
		outputPath, err := p.DocsOutputPath()
		if err != nil {
			return err
		}

		f, err := os.Create(outputPath)
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
