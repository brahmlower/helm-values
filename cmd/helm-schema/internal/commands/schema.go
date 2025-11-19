package commands

import (
	"helmschema/cmd/helm-schema/internal"
	"helmschema/cmd/helm-schema/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Schema(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewSchemaConfig()

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Generate values schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			return generateSchema(logger, cfg)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func generateSchema(logger *logrus.Logger, cfg *config.SchemaConfig) error {
	chartDir, err := cfg.ChartDir()
	if err != nil {
		return err
	}

	charts, err := internal.FindCharts(logger, chartDir)
	if err != nil {
		return err
	}
	logger.Infof("Found %d charts", len(charts))

	// Itterate through plan to set the logger and config
	plans := []*internal.Plan{}
	for _, chartRoot := range charts {
		plan := internal.NewPlan(
			chartRoot,
			cfg.StdOut(),
			cfg.Strict(),
			cfg.DryRun(),
		)
		plan.LogIntent(logger)
		plans = append(plans, plan)
	}

	// Iterate through plans again, this time generating the schema
	for _, plan := range plans {
		logger.Debugf("%s: schema: starting generation", plan.ChartRoot())
		schema, err := internal.NewGenerator(logger, plan).Generate()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		if err := plan.WriteSchema(schema); err != nil {
			logger.Error(err.Error())
			return nil
		}
	}

	return nil
}
