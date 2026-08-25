package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"helmvalues/pkg"
	"helmvalues/pkg/charts"

	"github.com/sirupsen/logrus"
)

// Plan holds the resolved config and target chart for a single schema generation run.
type Plan struct {
	cfg   *Config
	chart *charts.Chart
}

// NewPlan constructs a Plan combining the given schema config with the target chart.
func NewPlan(cfg *Config, chart *charts.Chart) *Plan {
	return &Plan{
		chart: chart,
		cfg:   cfg,
	}
}

// LogCommonDetails logs the config values shared across schema generation plans.
func (p *Plan) LogCommonDetails(logger *logrus.Logger) {
	// common configs
	logger.Debugf("plan: %s: DryRun=%t", p.chart.Details.Name, p.DryRun())
	logger.Debugf("plan: %s: StrictComments=%t", p.chart.Details.Name, p.StrictComments())
	logger.Debugf("plan: %s: Stdout=%t", p.chart.Details.Name, p.StdOut())
}

// LogChartDetails logs the resolved file paths for the plan's target chart.
func (p *Plan) LogChartDetails(logger *logrus.Logger) {
	// chart configs
	logger.Debugf("plan: %s: ChartRoot=%s", p.chart.Details.Name, p.chart.RootPath())
	logger.Debugf("plan: %s: ChartFile=%s", p.chart.Details.Name, p.chart.ChartFilePath())
	logger.Debugf("plan: %s: ChartValuesFile=%s", p.chart.Details.Name, p.chart.ValuesFilePath())
	logger.Debugf("plan: %s: ChartSchemaFile=%s", p.chart.Details.Name, p.chart.SchemaFilePath())
	// logger.Debugf("plan: %s: ChartReadmeTemplate=%s", p.chart.Details.Name, p.DocsChartReadmeTemplate())
}

// LogSchemaDetails logs the schema-specific config values for the plan.
func (p *Plan) LogSchemaDetails(logger *logrus.Logger) {
	logger.Debugf("plan: %s: WriteModeline=%t", p.chart.Details.Name, p.cfg.WriteModeline)
}

// Chart returns the plan's target chart.
func (p *Plan) Chart() *charts.Chart {
	return p.chart
}

// StdOut reports whether the generated schema should also be printed to stdout.
func (p *Plan) StdOut() bool {
	return p.cfg.StdOut
}

// StrictComments reports whether doc comment errors should be treated as fatal.
func (p *Plan) StrictComments() bool {
	return p.cfg.Strict
}

// GitAdd reports whether the generated schema file should be staged with git add.
func (p *Plan) GitAdd() bool {
	return p.cfg.GitAdd
}

// DryRun reports whether the plan should avoid writing any files.
func (p *Plan) DryRun() bool {
	return p.cfg.DryRun
}

// Check reports whether the plan should verify (rather than write) the
// chart's generated schema and Chart.yaml metadata.
func (p *Plan) Check() bool {
	return p.cfg.Check
}

// CheckSchema reports problems with the chart's on-disk schema and
// Chart.yaml metadata: whether values.schema.json matches the freshly
// generated schema, and whether the values-schema annotation references
// the chart's current version. An empty result means everything is up to
// date.
func (p *Plan) CheckSchema(schema *pkg.JsonSchema) []string {
	var problems []string

	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return []string{fmt.Sprintf("%s: failed to marshal generated schema: %s", p.chart.Details.Name, err)}
	}

	existing, err := os.ReadFile(p.chart.SchemaFilePath())
	if err != nil || !bytes.Equal(existing, content) {
		problems = append(problems, fmt.Sprintf(
			"%s is stale -- run `helm-values schema %s` and commit the result",
			p.chart.SchemaFilePath(), p.chart.RootPath(),
		))
	}

	if schemaURL := p.chart.Details.ValuesSchema(); schemaURL != "" {
		if !strings.Contains(schemaURL, p.chart.Details.Version) {
			problems = append(problems, fmt.Sprintf(
				"%s: values-schema annotation does not reference current chart version %s (found: %s)",
				p.chart.ChartFilePath(), p.chart.Details.Version, schemaURL,
			))
		}
	}

	return problems
}

// WriteSchema encodes and writes the generated schema per the plan's config.
func (p *Plan) WriteSchema(_ *logrus.Logger, schema *pkg.JsonSchema) error {
	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if p.StdOut() {
		fmt.Println(string(content))
	}

	if p.DryRun() {
		return nil
	}

	f, err := os.Create(p.chart.SchemaFilePath())
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	if p.GitAdd() {
		// The git add here is an intentional, user-opted-in integration (via the GitAdd
		// config flag) that stages the schema file this process just wrote; the arguments
		// are not attacker-controlled input.
		//nolint:gosec // opt-in git integration, not attacker input
		cmd := exec.CommandContext(context.Background(), "git", "add", p.chart.SchemaFilePath())
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to git add %s: %w", p.chart.SchemaFilePath(), err)
		}
	}

	return nil
}
