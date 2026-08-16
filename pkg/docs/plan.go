package docs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"helmvalues/pkg/charts"
	"helmvalues/pkg/docs/templates"
	"helmvalues/pkg/schema"

	"github.com/sirupsen/logrus"
)

// Plan holds the resolved configuration and chart-derived details needed to plan and
// generate documentation for a single chart.
type Plan struct {
	cfg        *Config
	chart      *charts.Chart
	schemaPlan *schema.Plan
}

// NewPlan creates a Plan that combines the given Config with a specific chart, deriving
// a schema.Plan for that chart along the way.
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

// LogCommonDetails logs the plan's common (non-chart-specific) configuration values.
func (p *Plan) LogCommonDetails(logger *logrus.Logger) {
	// common configs
	logger.Debugf("plan: %s: DryRun=%t", p.chart.Details.Name, p.DryRun())
	logger.Debugf("plan: %s: StrictComments=%t", p.chart.Details.Name, p.StrictComments())
	logger.Debugf("plan: %s: Stdout=%t", p.chart.Details.Name, p.StdOut())
}

// LogChartDetails logs the plan's chart-specific file path details.
func (p *Plan) LogChartDetails(logger *logrus.Logger) {
	// chart configs
	logger.Debugf("plan: %s: ChartRoot=%s", p.chart.Details.Name, p.chart.RootPath())
	logger.Debugf("plan: %s: ChartFile=%s", p.chart.Details.Name, p.chart.ChartFilePath())
	logger.Debugf("plan: %s: ChartValuesFile=%s", p.chart.Details.Name, p.chart.ValuesFilePath())
	logger.Debugf("plan: %s: ChartSchemaFile=%s", p.chart.Details.Name, p.chart.SchemaFilePath())
	logger.Debugf("plan: %s: ChartReadmeTemplate=%s", p.chart.Details.Name, p.DocsChartReadmeTemplate())
}

// LogDocDetails logs the plan's resolved documentation generation details, such as the
// target template, markup type, and output path.
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

// LogSchemaDetails logs the plan's underlying schema.Plan details.
func (p *Plan) LogSchemaDetails(logger *logrus.Logger) {
	p.schemaPlan.LogSchemaDetails(logger)
}

// Chart returns the chart this plan generates documentation for.
func (p *Plan) Chart() *charts.Chart {
	return p.chart
}

// StdOut reports whether the generated documentation should also be printed to stdout.
func (p *Plan) StdOut() bool {
	return p.cfg.StdOut
}

// StrictComments reports whether strict comment validation is enabled.
func (p *Plan) StrictComments() bool {
	return p.cfg.Strict
}

// GitAdd reports whether the generated output file should be staged with `git add`.
func (p *Plan) GitAdd() bool {
	return p.cfg.GitAdd
}

// DryRun reports whether the plan should skip writing output to disk.
func (p *Plan) DryRun() bool {
	return p.cfg.DryRun
}

// DocsTargetTemplate resolves the template path to render, returning whether the
// built-in default template should be used, or an error if no template can be found.
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

// DocsChartReadmeTemplate returns the path of the chart's own README template file
// (Markdown or reStructuredText), or an empty string if the chart has none.
func (p *Plan) DocsChartReadmeTemplate() string {
	if _, err := os.Stat(p.chart.ReadmeMdTemplateFilePath()); err == nil {
		return p.chart.ReadmeMdTemplateFilePath()
	}

	if _, err := os.Stat(p.chart.ReadmeRstTemplateFilePath()); err == nil {
		return p.chart.ReadmeRstTemplateFilePath()
	}

	return ""
}

// DocsMarkup resolves the markup type to render, inferring it from the configured
// markup, template path, or chart readme template when not explicitly set.
func (p *Plan) DocsMarkup() (templates.Markup, error) {
	if value, ok := p.cfg.Markup.Get(); ok {
		return value, nil
	}

	if p.DocsUseDefault() {
		return templates.Markdown, nil
	}

	// If a template was specified, infer the markup type from that
	if tmpl := p.cfg.Template; tmpl != "" {
		markup, err := templates.MarkupFromPath(tmpl)
		if err != nil {
			return "", fmt.Errorf("inferring markup from template path %s: %w", tmpl, err)
		}

		return markup, nil
	}

	// If there's a readme template in the chart, infer the markup type from that
	if tmpl := p.DocsChartReadmeTemplate(); tmpl != "" {
		markup, err := templates.MarkupFromPath(tmpl)
		if err != nil {
			return "", fmt.Errorf("inferring markup from chart readme template path %s: %w", tmpl, err)
		}

		return markup, nil
	}

	return "", errors.New("unable to infer markup type")
}

// DocsUseDefault reports whether the built-in default template should be used, based on
// the explicit config value, a configured custom template, or a chart-provided template.
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

// DocsOutputPath resolves the file path the generated documentation should be written
// to, using the configured output path or deriving one from the chart and markup type.
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

// SchemaPlan returns the plan's underlying schema.Plan.
func (p *Plan) SchemaPlan() *schema.Plan {
	return p.schemaPlan
}

// WriteReadme writes the rendered documentation content to the plan's output path (unless
// DryRun is set), and/or prints it to stdout, depending on the plan's configuration.
func (p *Plan) WriteReadme(logger *logrus.Logger, content string) error {
	if !p.DryRun() {
		logger.Debugf("plan: %s: writing readme to disk", p.chart.Details.Name)

		if err := p.writeReadmeFile(content); err != nil {
			return err
		}
	}

	if p.StdOut() {
		fmt.Println(content)
	}

	return nil
}

// writeReadmeFile writes content to the plan's output path, and, if configured,
// stages the resulting file with `git add`.
func (p *Plan) writeReadmeFile(content string) error {
	outputPath, err := p.DocsOutputPath()
	if err != nil {
		return err
	}

	// outputPath is derived from the chart's own configured/derived README path
	// (or a user-supplied --output flag), not attacker-controlled input.
	//nolint:gosec // outputPath is a user-opted-in, non-attacker-controlled file path
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err = f.WriteString(content); err != nil {
		return fmt.Errorf("writing output file %s: %w", outputPath, err)
	}

	if p.GitAdd() {
		// outputPath is derived from the chart's own configured/derived README path
		// (or a user-supplied --output flag), not attacker-controlled input; running
		// `git add` on it is an intentional, user-opted-in git integration.
		//nolint:gosec // outputPath is a user-opted-in, non-attacker-controlled file path
		err := exec.CommandContext(context.Background(), "git", "add", outputPath).Run()
		if err != nil {
			return fmt.Errorf("failed to git add %s: %w", outputPath, err)
		}
	}

	return nil
}
