package config

import (
	"fmt"

	"helmvalues/pkg/schema"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SchemaConfig holds the flag/env-bound configuration for the schema
// command.
type SchemaConfig struct {
	*viper.Viper
}

// NewSchemaConfig creates a SchemaConfig backed by a fresh viper instance.
func NewSchemaConfig() *SchemaConfig {
	cfg := standardViper()

	return &SchemaConfig{cfg}
}

// LogLevel returns the configured log level.
func (c *SchemaConfig) LogLevel() (logrus.Level, error) {
	level, err := logrus.ParseLevel(c.GetString(logLevelFlag))
	if err != nil {
		return level, fmt.Errorf("parsing log level: %w", err)
	}

	return level, nil
}

// UpdateLogger sets logger's level to the configured log level.
func (c *SchemaConfig) UpdateLogger(logger *logrus.Logger) error {
	level, err := c.LogLevel()
	if err != nil {
		return err
	}

	logger.SetLevel(level)

	return nil
}

// BindFlags registers the schema command's flags on cmd and binds them
// (and their environment-variable equivalents) to this config.
func (c *SchemaConfig) BindFlags(cmd *cobra.Command) error {
	cmd.Flags().Bool("stdout", false, "write to stdout")
	cmd.Flags().Bool("strict", false, "fail on doc comment parsing errors")
	cmd.Flags().Bool("git-add", false, "stage changes with git add (useful for pre-commit hooks)")
	cmd.Flags().Bool("dry-run", false, "don't write changes to disk")
	cmd.Flags().String(logLevelFlag, "warn", "log level (debug, info, warn, error, fatal, panic)")
	cmd.Flags().Bool("write-modeline", true, "write modeline to values file")
	cmd.Flags().Bool("check", false,
		"check that values.schema.json is up to date and that Chart.yaml's "+
			"values-schema annotation references the chart's current version, "+
			"without writing changes (exit non-zero if not)")

	for _, name := range []string{
		"stdout", "strict", "git-add", "dry-run", logLevelFlag, "write-modeline", "check",
	} {
		if err := bindFlag(c.Viper, cmd, name); err != nil {
			return err
		}
	}

	return nil
}

// ToPackageConfig builds the schema.Config this configuration describes.
func (c *SchemaConfig) ToPackageConfig() (*schema.Config, error) {
	logLevel, err := c.LogLevel()
	if err != nil {
		return nil, err
	}

	config := &schema.Config{
		StdOut:        c.GetBool("stdout"),
		Strict:        c.GetBool("strict"),
		DryRun:        c.GetBool("dry-run"),
		GitAdd:        c.GetBool("git-add"),
		WriteModeline: c.GetBool("write-modeline"),
		Check:         c.GetBool("check"),
		LogLevel:      logLevel,
	}

	return config, nil
}
