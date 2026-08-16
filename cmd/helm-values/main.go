// Command helm-values is a Helm plugin that generates a JSON Schema and
// documentation for a chart's values.yaml, and manages the
// yaml-language-server modeline and pre-commit hooks that go with them.
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
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return fmt.Errorf("updating logger: %w", err)
			}

			schemaCfg, err := cfg.ToPackageConfig()
			if err != nil {
				return fmt.Errorf("building schema config: %w", err)
			}

			return schema.GenerateSchema(logger, schemaCfg, args)
		},
		GroupID: group.ID,
	}

	if err := cfg.BindFlags(cmd); err != nil {
		panic(fmt.Sprintf("binding schema flags: %v", err))
	}

	return cmd
}

func CommandDocs(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
	cfg := config.NewDocsConfig()

	cmd := &cobra.Command{
		Use:   "docs [flags] chart_dir [...chart_dir]",
		Short: "Generate values docs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return fmt.Errorf("updating logger: %w", err)
			}

			docsCfg, err := cfg.ToPackageConfig()
			if err != nil {
				return fmt.Errorf("building docs config: %w", err)
			}

			return docs.GenerateDocs(logger, docsCfg, args)
		},
		GroupID: group.ID,
	}

	if err := cfg.BindFlags(cmd); err != nil {
		panic(fmt.Sprintf("binding docs flags: %v", err))
	}

	return cmd
}

// modelineMaxArgs is the maximum number of positional args the modeline
// command accepts: chart_ref and, optionally, values_file.
const modelineMaxArgs = 2

func CommandModeline(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
	cfg := config.NewModelineConfig()

	cmd := &cobra.Command{
		Use:     "modeline [flags] chart_ref values_file",
		Short:   "Add yaml-language-server modeline to values file",
		GroupID: group.ID,
		Args:    cobra.RangeArgs(1, modelineMaxArgs),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return fmt.Errorf("updating logger: %w", err)
			}

			chartRef := args[0]

			valuesFile := ""
			if len(args) > 1 {
				valuesFile = args[1]
			}

			modelineCfg, err := cfg.ToPackageConfig(chartRef, valuesFile)
			if err != nil {
				return fmt.Errorf("building modeline config: %w", err)
			}

			return modeline.WriteModeline(logger, modelineCfg)
		},
	}

	if err := cfg.BindFlags(cmd); err != nil {
		panic(fmt.Sprintf("binding modeline flags: %v", err))
	}

	return cmd
}

func CommandUpdate(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the helm-values plugin to the latest version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return internal.Update(cmd.Context(), logger, Repository, BuildVersion)
		},
	}

	return cmd
}

func CommandPreCommit(logger *logrus.Logger, group *cobra.Group) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pre-commit",
		Short:   "Install pre-commit hooks for generating schema and docs",
		GroupID: group.ID,
		RunE: func(_ *cobra.Command, _ []string) error {
			return internal.InstallPreCommitHooks(logger)
		},
	}

	return cmd
}

func CommandVersion(_ *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(_ *cobra.Command, _ []string) error {
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
