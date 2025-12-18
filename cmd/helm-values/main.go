package main

import (
	"os"

	"helmschema/cmd/helm-values/internal/config"
	"helmschema/pkg/docs"
	"helmschema/pkg/schema"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	err := Program(logger).Execute()
	if err != nil {
		os.Exit(1)
	}
}

func Program(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm-values",
		Short: "Schema and docs generator for Helm values",
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(Schema(logger))
	cmd.AddCommand(Docs(logger))
	return cmd
}

func Schema(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewSchemaConfig()

	cmd := &cobra.Command{
		Use:   "schema [flags] chart_dir [...chart_dir]",
		Short: "Generate values schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			schemaCfg, err := cfg.ToPackageConfig()
			if err != nil {
				return err
			}
			return schema.GenerateSchema(logger, schemaCfg, args)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func Docs(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewDocsConfig()

	cmd := &cobra.Command{
		Use:   "docs [flags] chart_dir [...chart_dir]",
		Short: "Generate values docs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			docsCfg, err := cfg.ToPackageConfig()
			if err != nil {
				return err
			}
			return docs.GenerateDocs(logger, docsCfg, args)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}
