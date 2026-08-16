package schema

import (
	"fmt"

	"helmvalues/pkg/charts"

	"github.com/sirupsen/logrus"
)

// GenerateSchema searches chartDirs for charts and generates a JSON schema (and,
// depending on cfg, a values file modeline) for each one found.
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
	for _, plan := range plans {
		logger.Infof("schema: %s: starting generation", plan.Chart().Details.Name)

		schema, err := NewGenerator(logger, plan).Generate()
		if err != nil {
			logger.Error(err.Error())

			return nil
		}

		logger.Debugf("schema: %s: writing output", plan.Chart().Details.Name)

		if err := plan.WriteSchema(logger, schema); err != nil {
			logger.Error(err.Error())

			return nil
		}

		if cfg.WriteModeline {
			logger.Debugf("schema: %s: writing modeline", plan.Chart().Details.Name)

			err := WriteSchemaModeline(
				logger,
				plan.Chart(),
				plan.Chart().ValuesFilePath(),
				plan.DryRun(),
			)
			if err != nil {
				logger.Error(err.Error())

				return nil
			}
		} else {
			logger.Debugf("schema: %s: skipping modeline write", plan.Chart().Details.Name)
		}

		logger.Infof("schema: %s: finished", plan.Chart().Details.Name)
	}

	return nil
}
