package commands

import (
	"helmschema/cmd/helm-values/internal"
	"helmschema/cmd/helm-values/internal/charts"
	"helmschema/cmd/helm-values/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Schema(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewSchemaConfig()

	cmd := &cobra.Command{
		Use:   "schema [flags] chart_dir [...chart_dir]",
		Short: "Generate values schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			return generateSchema(logger, cfg, args)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func generateSchema(logger *logrus.Logger, cfg *config.SchemaConfig, chartDirs []string) error {
	chartsFound, err := charts.Search(logger, chartDirs)
	if err != nil {
		return err
	}

	// Itterate through plan to set the logger and config
	plans := []*internal.Plan{}
	for _, chart := range chartsFound {
		plan := internal.NewPlan(nil, cfg, chart)
		plan.LogIntent(logger)
		plans = append(plans, plan)
	}

	// Iterate through plans again, this time generating the schema
	for _, plan := range plans {
		logger.Infof("schema: %s: starting generation", plan.Chart().Details.Name)
		schema, err := internal.NewGenerator(logger, plan).Generate()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		logger.Debugf("schema: %s: writing output", plan.Chart().Details.Name)
		if err := plan.WriteSchema(logger, schema); err != nil {
			logger.Error(err.Error())
			return nil
		}

		if cfg.WriteModeline() {
			logger.Debugf("schema: %s: writing modeline", plan.Chart().Details.Name)
			if err := plan.WriteSchemaModeline(logger); err != nil {
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
