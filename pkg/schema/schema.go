package schema

import (
	"helmvalues/internal/charts"

	"github.com/sirupsen/logrus"
)

func GenerateSchema(logger *logrus.Logger, cfg *Config, chartDirs []string) error {
	chartsFound, err := charts.Search(logger, chartDirs)
	if err != nil {
		return err
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
			if err := WriteSchemaModeline(logger, plan.Chart(), plan.DryRun()); err != nil {
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
