package config

import (
	"helmvalues/pkg/docs"
	"helmvalues/pkg/docs/templates"
	"path/filepath"

	"github.com/samber/mo"
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

func (c *DocsConfig) ValuesOrder() (docs.ValuesOrder, error) {
	return docs.NewValuesOrder(c.GetString("order"))
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

func (c *DocsConfig) Markup() (mo.Option[templates.Markup], error) {
	if !c.IsSet("markup") {
		return mo.None[templates.Markup](), nil
	}
	markup, err := templates.MarkupFromString(c.GetString("markup"))
	if err != nil {
		return mo.None[templates.Markup](), err
	}
	return mo.Some(markup), nil
}

func (c *DocsConfig) UseDefault() mo.Option[bool] {
	if !c.IsSet("use-default") {
		return mo.None[bool]()
	}
	return mo.Some(c.GetBool("use-default"))
}

func (c *DocsConfig) Output() mo.Option[string] {
	if !c.IsSet("output") {
		return mo.None[string]()
	}
	return mo.Some(c.GetString("output"))
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

	cmd.Flags().String("markup", "", "markup language (md, markdown, rst, restructuredtext)")
	c.BindPFlag("markup", cmd.Flags().Lookup("markup"))
	c.BindEnv("markup")

	cmd.Flags().String("order", "preserve", "order of values (preserve, alphabetical)")
	c.BindPFlag("order", cmd.Flags().Lookup("order"))
	c.BindEnv("order")

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

func (c *DocsConfig) ToPackageConfig() (*docs.Config, error) {
	logLevel, err := c.LogLevel()
	if err != nil {
		return nil, err
	}

	extraTemplates, err := c.ExtraTemplates()
	if err != nil {
		return nil, err
	}

	valuesOrder, err := c.ValuesOrder()
	if err != nil {
		return nil, err
	}

	markup, err := c.Markup()
	if err != nil {
		return nil, err
	}

	config := &docs.Config{
		LogLevel:       logLevel,
		StdOut:         c.GetBool("stdout"),
		Strict:         c.GetBool("strict"),
		DryRun:         c.GetBool("dry-run"),
		UseDefault:     c.UseDefault(),
		Output:         c.Output(),
		Template:       c.GetString("template"),
		ExtraTemplates: extraTemplates,
		Markup:         markup,
		Order:          valuesOrder,
	}
	return config, nil
}
