package commands

import (
	"helmschema/cmd/helm-schema/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Docs(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewDocsConfig()

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate values docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			return generateDocs(logger, cfg)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func generateDocs(logger *logrus.Logger, cfg *config.DocsConfig) error {
	return nil
}
