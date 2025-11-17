package main

import (
	"helmschema/cmd/helm-schema/internal"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GenerateCommand(logger *logrus.Logger) *cobra.Command {
	v := viper.New()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the worker server",
		RunE: func(cmd *cobra.Command, args []string) error {
			level, err := logrus.ParseLevel(v.GetString("log-level"))
			if err != nil {
				return err
			}
			logger.SetLevel(level)

			chartDir, err := filepath.Abs(v.GetString("chart-dir"))
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
				plan.Stdout = v.GetBool("stdout")
				plan.StrictComments = v.GetBool("strict")
				plan.DryRun = v.GetBool("dry-run")
				plan.SetSchemaFilename(v.GetString("schema-file"))

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
		},
	}

	cmd.Flags().String("chart-dir", "", "path to the chart directory")
	v.BindPFlag("chart-dir", cmd.Flags().Lookup("chart-dir"))
	v.BindEnv("chart-dir")

	cmd.Flags().String("schema-file", "values.schema.json", "path to the schema-file file")
	v.BindPFlag("schema-file", cmd.Flags().Lookup("schema-file"))
	v.BindEnv("schema-file")

	cmd.Flags().Bool("stdout", false, "write to stdout")
	v.BindPFlag("stdout", cmd.Flags().Lookup("stdout"))
	v.BindEnv("stdout")

	cmd.Flags().Bool("strict", false, "fail on doc comment parsing errors")
	v.BindPFlag("strict", cmd.Flags().Lookup("strict"))
	v.BindEnv("strict")

	cmd.Flags().Bool("dry-run", false, "don't write changes to disk")
	v.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))
	v.BindEnv("dry-run")

	cmd.Flags().String("log-level", "warn", "log level (debug, info, warn, error, fatal, panic)")
	v.BindPFlag("log-level", cmd.Flags().Lookup("log-level"))
	v.BindEnv("log-level")

	return cmd
}
