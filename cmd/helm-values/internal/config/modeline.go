package config

import (
	"fmt"

	"helmvalues/pkg/helm"
	"helmvalues/pkg/modeline"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ModelineConfig holds the flag/env-bound configuration for the modeline
// command.
type ModelineConfig struct {
	*viper.Viper
}

// NewModelineConfig creates a ModelineConfig backed by a fresh viper
// instance.
func NewModelineConfig() *ModelineConfig {
	cfg := standardViper()

	return &ModelineConfig{cfg}
}

// LogLevel returns the configured log level.
func (c *ModelineConfig) LogLevel() (logrus.Level, error) {
	level, err := logrus.ParseLevel(c.GetString(logLevelFlag))
	if err != nil {
		return level, fmt.Errorf("parsing log level: %w", err)
	}

	return level, nil
}

// UpdateLogger sets logger's level to the configured log level.
func (c *ModelineConfig) UpdateLogger(logger *logrus.Logger) error {
	level, err := c.LogLevel()
	if err != nil {
		return err
	}

	logger.SetLevel(level)

	return nil
}

// BindFlags registers the modeline command's flags on cmd and binds them
// (and their environment-variable equivalents) to this config.
func (c *ModelineConfig) BindFlags(cmd *cobra.Command) error {
	cmd.Flags().BoolP("parents", "p", false, "create parent directories if they don't exist")
	cmd.Flags().String("version", "", "chart version (for remote charts)")
	cmd.Flags().String(logLevelFlag, "warn", "log level (debug, info, warn, error, fatal, panic)")

	for _, name := range []string{"parents", "version", logLevelFlag} {
		if err := bindFlag(c.Viper, cmd, name); err != nil {
			return err
		}
	}

	return nil
}

// ToPackageConfig builds the modeline.Config this configuration describes
// for the given chart reference and target file.
func (c *ModelineConfig) ToPackageConfig(rawChartRef string, targetFile string) (*modeline.Config, error) {
	if version := c.GetString("version"); version != "" {
		rawChartRef = rawChartRef + "@" + version
	}

	chartRef, err := helm.NewChartRef(rawChartRef)
	if err != nil {
		return nil, fmt.Errorf("parsing chart reference: %w", err)
	}

	modelineCfg := &modeline.Config{
		ChartRef:        chartRef,
		TargetFile:      targetFile,
		CreateParents:   c.GetBool("parents"),
		PartialModeline: modeline.NewPartialModeline("yaml-language-server", "$schema"),
	}

	return modelineCfg, nil
}
