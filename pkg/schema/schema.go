package schema

import (
	"errors"
	"fmt"
	"os"

	"helmvalues/pkg/charts"

	"github.com/sirupsen/logrus"
)

// GenerateSchema searches chartDirs for charts and generates a JSON schema (and,
// depending on cfg, a values file modeline) for each one found. A chart that
// fails doesn't stop the others: every chart is processed, and the errors from
// failing charts are joined with any check problems into the returned error.
func GenerateSchema(logger *logrus.Logger, cfg *Config, chartDirs []string) error {
	chartsFound, err := charts.Search(logger, chartDirs)
	if err != nil {
		return fmt.Errorf("failed to search charts: %w", err)
	}

	// Itterate through plan to set the logger and config
	plans := []*Plan{}

	for _, chart := range chartsFound {
		plan := NewPlan(cfg, chart)
		plan.LogCommonDetails(logger)
		plan.LogChartDetails(logger)
		plan.LogSchemaDetails(logger)

		plans = append(plans, plan)
	}

	// Iterate through plans again, this time generating the schema
	var (
		problems  []string
		chartErrs []error
	)

	for _, plan := range plans {
		name := plan.Chart().Details.Name

		logger.Infof("schema: %s: starting generation", name)

		chartProblems, err := generateChartSchema(logger, cfg, plan)
		if err != nil {
			chartErrs = append(chartErrs, fmt.Errorf("schema: %s: %w", name, err))

			continue
		}

		problems = append(problems, chartProblems...)
	}

	return errors.Join(append(chartErrs, reportProblems(problems))...)
}

// generateChartSchema generates the schema for a single chart's plan, then
// either checks it against what's on disk or writes it (and, depending on cfg,
// the values file modeline). The returned problems are non-empty only when
// plan.Check() is set and the generated schema doesn't match what's on disk.
func generateChartSchema(logger *logrus.Logger, cfg *Config, plan *Plan) ([]string, error) {
	name := plan.Chart().Details.Name

	schema, err := NewGenerator(logger, plan).Generate()
	if err != nil {
		return nil, err
	}

	if plan.Check() {
		return plan.CheckSchema(schema), nil
	}

	logger.Debugf("schema: %s: writing output", name)

	if err := plan.WriteSchema(logger, schema); err != nil {
		return nil, err
	}

	if !cfg.WriteModeline {
		logger.Debugf("schema: %s: skipping modeline write", name)
		logger.Infof("schema: %s: finished", name)

		return nil, nil
	}

	logger.Debugf("schema: %s: writing modeline", name)

	err = WriteSchemaModeline(logger, plan.Chart(), plan.Chart().ValuesFilePath(), plan.DryRun())
	if err != nil {
		return nil, err
	}

	logger.Infof("schema: %s: finished", name)

	return nil, nil
}

// reportProblems prints every check problem found and, if any were found,
// returns an error summarizing how many.
func reportProblems(problems []string) error {
	if len(problems) == 0 {
		return nil
	}

	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p)
	}

	return fmt.Errorf("check failed: %d problem(s) found", len(problems))
}
