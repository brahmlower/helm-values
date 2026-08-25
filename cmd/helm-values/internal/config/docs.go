package config

import (
	"fmt"
	"path/filepath"

	"helmvalues/pkg/docs"
	"helmvalues/pkg/docs/templates"

	"github.com/samber/mo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DocsConfig holds the flag/env-bound configuration for the docs command.
type DocsConfig struct {
	*viper.Viper
}

// NewDocsConfig creates a DocsConfig backed by a fresh viper instance.
func NewDocsConfig() *DocsConfig {
	cfg := standardViper()

	return &DocsConfig{cfg}
}

// ValuesOrder returns the configured order in which values rows are
// rendered.
func (c *DocsConfig) ValuesOrder() (docs.ValuesOrder, error) {
	order, err := docs.NewValuesOrder(c.GetString("order"))
	if err != nil {
		return order, fmt.Errorf("parsing values order: %w", err)
	}

	return order, nil
}

// LogLevel returns the configured log level.
func (c *DocsConfig) LogLevel() (logrus.Level, error) {
	level, err := logrus.ParseLevel(c.GetString(logLevelFlag))
	if err != nil {
		return level, fmt.Errorf("parsing log level: %w", err)
	}

	return level, nil
}

// ExtraTemplates resolves the configured extra-templates glob into a list
// of matching file paths.
func (c *DocsConfig) ExtraTemplates() ([]string, error) {
	et := c.GetString("extra-templates")
	if et == "" {
		return nil, nil
	}

	path, err := filepath.Abs(et)
	if err != nil {
		return nil, fmt.Errorf("resolving extra-templates path: %w", err)
	}

	matches, err := filepath.Glob(path)
	if err != nil {
		return nil, fmt.Errorf("globbing extra-templates path: %w", err)
	}

	return matches, nil
}

// Markup returns the configured output markup, if one was set.
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

// UseDefault returns the configured use-default flag, if it was set.
func (c *DocsConfig) UseDefault() mo.Option[bool] {
	if !c.IsSet("use-default") {
		return mo.None[bool]()
	}

	return mo.Some(c.GetBool("use-default"))
}

// Output returns the configured output path, if one was set.
func (c *DocsConfig) Output() mo.Option[string] {
	if !c.IsSet("output") {
		return mo.None[string]()
	}

	return mo.Some(c.GetString("output"))
}

// UpdateLogger sets logger's level to the configured log level.
func (c *DocsConfig) UpdateLogger(logger *logrus.Logger) error {
	level, err := c.LogLevel()
	if err != nil {
		return err
	}

	logger.SetLevel(level)

	return nil
}

// BindFlags registers the docs command's flags on cmd and binds them (and
// their environment-variable equivalents) to this config.
func (c *DocsConfig) BindFlags(cmd *cobra.Command) error {
	cmd.Flags().Bool("stdout", false, "write to stdout")
	cmd.Flags().Bool("git-add", false, "stage changes with git add (useful for pre-commit hooks)")
	cmd.Flags().Bool("strict", false, "fail on doc comment parsing errors")
	cmd.Flags().Bool("dry-run", false, "don't write changes to disk")
	cmd.Flags().String(logLevelFlag, "warn", "log level (debug, info, warn, error, fatal, panic)")
	cmd.Flags().String("markup", "", "markup language (md, markdown, rst, restructuredtext)")
	cmd.Flags().String("order", "preserve", "order of values (preserve, alphabetical)")
	cmd.Flags().Bool("use-default", true, "uses default template unless a custom template is present")
	cmd.Flags().String("output", "", "path to output (defaults to README.md or README.rst based on markup)")
	cmd.Flags().String("template", "", "path to template (defaults to README.md.tmpl or README.rst.tmpl based on markup)")
	cmd.Flags().String("extra-templates", "", "glob path to extra templates")
	cmd.Flags().Bool("check", false,
		"check that the rendered docs file is up to date, without writing "+
			"changes (exit non-zero if not)")

	for _, name := range []string{
		"stdout", "git-add", "strict", "dry-run", logLevelFlag, "markup",
		"order", "use-default", "output", "template", "extra-templates", "check",
	} {
		if err := bindFlag(c.Viper, cmd, name); err != nil {
			return err
		}
	}

	return nil
}

// ToPackageConfig builds the docs.Config this configuration describes.
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
		GitAdd:         c.GetBool("git-add"),
		Check:          c.GetBool("check"),
		UseDefault:     c.UseDefault(),
		Output:         c.Output(),
		Template:       c.GetString("template"),
		ExtraTemplates: extraTemplates,
		Markup:         markup,
		Order:          valuesOrder,
	}

	return config, nil
}
