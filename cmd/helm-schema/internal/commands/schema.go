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

	plans, err := internal.BuildPlan(chartDir, logger)
	if err != nil {
		return err
	}
	logger.Infof("Found %d charts", len(plans))

	// Itterate through plan to set the logger and config
	for _, plan := range plans {
		plan.Logger = logger
		plan.Stdout = cfg.StdOut()
		plan.StrictComments = cfg.Strict()
		plan.DryRun = cfg.DryRun()
		plan.SetSchemaFilename(cfg.SchemaFile())

		plan.LogIntent()
	}

	// Iterate through plans again, this time generating the schema
	for _, plan := range plans {
		logger.Debugf("%s: schema: starting generation", plan.ChartDir)
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
