package main

import (
	"fmt"
	"os"
	"strings"

	"helmvalues/cmd/helm-values/internal"
	"helmvalues/cmd/helm-values/internal/config"
	"helmvalues/pkg/docs"
	"helmvalues/pkg/modeline"
	"helmvalues/pkg/schema"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var BuildVersion string
var BuildCommit string
var BuildDate string
var Repository string

var generationGroup = &cobra.Group{
	ID:    "generation",
	Title: "Generation Commands:",
}
var utilityGroup = &cobra.Group{
	ID:    "utility",
	Title: "Utility Commands:",
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

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

	cmd.AddGroup(generationGroup)
	cmd.AddGroup(utilityGroup)

	cmd.AddCommand(CommandSchema(logger, generationGroup))
	cmd.AddCommand(CommandDocs(logger, generationGroup))
	cmd.AddCommand(CommandPreCommit(logger, utilityGroup))
	cmd.AddCommand(CommandModeline(logger, utilityGroup))
	cmd.AddCommand(CommandUpdate(logger))
	cmd.AddCommand(CommandVersion(logger))
	return cmd
}

func CommandSchema(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
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
		GroupID: group.ID,
	}

	cfg.BindFlags(cmd)

	return cmd
}

func CommandDocs(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
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
		GroupID: group.ID,
	}

	cfg.BindFlags(cmd)

	return cmd
}

func CommandModeline(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
	cfg := config.NewModelineConfig()

	cmd := &cobra.Command{
		Use:     "modeline [flags] chart_ref values_file",
		Short:   "Add yaml-language-server modeline to values file",
		GroupID: group.ID,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			chartRef := args[0]
			valuesFile := ""
			if len(args) > 1 {
				valuesFile = args[1]
			}

			modelineCfg, err := cfg.ToPackageConfig(chartRef, valuesFile)
			if err != nil {
				return err
			}

			return modeline.WriteModeline(logger, modelineCfg)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func CommandUpdate(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the helm-values plugin to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.Update(logger, Repository, BuildVersion)
		},
	}

	return cmd
}

func CommandPreCommit(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pre-commit",
		Short:   "Install pre-commit hooks for generating schema and docs",
		GroupID: group.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.InstallPreCommitHooks(logger)
		},
	}

	return cmd
}

func CommandVersion(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseNotes := ""

			if !strings.Contains(BuildVersion, "SNAPSHOT") {
				releaseNotes = fmt.Sprintf("https://github.com/%s/releases/tag/%s", Repository, BuildVersion)
			}

			fmt.Printf("helm-values\n\n")
			fmt.Printf("  Version:       %s\n", BuildVersion)
			fmt.Printf("  Commit:        %s\n", BuildCommit)
			fmt.Printf("  Date:          %s\n", BuildDate)
			fmt.Printf("  Release Notes: %s\n", releaseNotes)
			return nil
		},
	}

	return cmd
}
