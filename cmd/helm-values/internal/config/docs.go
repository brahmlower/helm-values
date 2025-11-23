package config

import (
	"helmschema/cmd/helm-values/internal/docs"
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
	et := c.GetString("extra-templates")
	if et == "" {
		return nil, nil
	}

	path, err := filepath.Abs(et)
	if err != nil {
		return nil, err
	}

	return filepath.Glob(path)
}

func (c *DocsConfig) Template() string {
	return c.GetString("template")
}

func (c *DocsConfig) Markup() (docs.Markup, bool, error) {
	if !c.IsSet("markup") {
		return "", false, nil
	}
	markup, err := docs.MarkupFromString(c.GetString("markup"))
	if err != nil {
		return "", true, err
	}
	return markup, true, nil
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

	cmd.Flags().String("markup", "markdown", "markup language (markdown, restructuredtext)")
	c.BindPFlag("markup", cmd.Flags().Lookup("markup"))
	c.BindEnv("markup")

	cmd.Flags().Bool("use-default", true, "uses default template unless a custom template is present")
	c.BindPFlag("use-default", cmd.Flags().Lookup("use-default"))
	c.BindEnv("use-default")

	cmd.Flags().String("output", "", "path to output (defaults to README.md or README.rst based on markup)")
	c.BindPFlag("output", cmd.Flags().Lookup("output"))
	c.BindEnv("output")

	cmd.Flags().String("template", "", "path to template (defaults to README.md.tmpl or README.rst.tmpl based on markup)")
	c.BindPFlag("template", cmd.Flags().Lookup("template"))
	c.BindEnv("template")

	cmd.Flags().String("extra-templates", "", "glob path to extra templates")
	c.BindPFlag("extra-templates", cmd.Flags().Lookup("extra-templates"))
	c.BindEnv("extra-templates")
}
