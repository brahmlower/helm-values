package main

import (
	"os"

	"helmschema/cmd/helm-schema/internal/commands"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Program(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm-values",
		Short: "Schema and docs generator for Helm values",
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(commands.Schema(logger))
	cmd.AddCommand(commands.Docs(logger))
	return cmd
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	err := Program(logger).Execute()
	if err != nil {
		os.Exit(1)
	}
}
