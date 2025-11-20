package config

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDocsConfig() *DocsConfig {
	cfg := standardViper()

	return &DocsConfig{cfg}
}

type DocsConfig struct {
	*viper.Viper
}

func (c *DocsConfig) ChartDir() (string, error) {
	return filepath.Abs(c.GetString("chart-dir"))
}

func (c *DocsConfig) SchemaFile() string {
	return c.GetString("schema-file")
}

func (c *DocsConfig) StdOut() bool {
	return c.GetBool("stdout")
}

func (c *DocsConfig) Strict() bool {
	return c.GetBool("strict")
}

func (c *DocsConfig) DryRun() bool {
	return c.GetBool("dry-run")
}

func (c *DocsConfig) LogLevel() (logrus.Level, error) {
	return logrus.ParseLevel(c.GetString("log-level"))
}

func (c *DocsConfig) ExtraTemplates() ([]string, error) {
	path, err := filepath.Abs(c.GetString("extra-templates"))
	if err != nil {
		return nil, err
	}

	return filepath.Glob(path)
}

func (c *DocsConfig) UpdateLogger(logger *logrus.Logger) error {
	level, err := c.LogLevel()
	if err != nil {
		return err
	}

	logger.SetLevel(level)
	return nil
}

func (c *DocsConfig) BindFlags(cmd *cobra.Command) {
	cmd.Flags().String("chart-dir", "", "path to the chart directory")
	c.BindPFlag("chart-dir", cmd.Flags().Lookup("chart-dir"))
	c.BindEnv("chart-dir")

	cmd.Flags().String("schema-file", "values.schema.json", "path to the schema-file file")
	c.BindPFlag("schema-file", cmd.Flags().Lookup("schema-file"))
	c.BindEnv("schema-file")

	cmd.Flags().Bool("stdout", false, "write to stdout")
	c.BindPFlag("stdout", cmd.Flags().Lookup("stdout"))
	c.BindEnv("stdout")

	cmd.Flags().Bool("strict", false, "fail on doc comment parsing errors")
	c.BindPFlag("strict", cmd.Flags().Lookup("strict"))
	c.BindEnv("strict")

	cmd.Flags().Bool("dry-run", false, "don't write changes to disk")
	c.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))
	c.BindEnv("dry-run")

	cmd.Flags().String("log-level", "warn", "log level (debug, info, warn, error, fatal, panic)")
	c.BindPFlag("log-level", cmd.Flags().Lookup("log-level"))
	c.BindEnv("log-level")

	cmd.Flags().String("extra-templates", "", "path to extra templates directory")
	c.BindPFlag("extra-templates", cmd.Flags().Lookup("extra-templates"))
	c.BindEnv("extra-templates")
}
