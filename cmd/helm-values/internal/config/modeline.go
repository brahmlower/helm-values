package config

import (
	"helmvalues/pkg/modeline"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewModelineConfig() *ModelineConfig {
	cfg := standardViper()

	return &ModelineConfig{cfg}
}

type ModelineConfig struct {
	*viper.Viper
}

func (c *ModelineConfig) LogLevel() (logrus.Level, error) {
	return logrus.ParseLevel(c.GetString("log-level"))
}

func (c *ModelineConfig) UpdateLogger(logger *logrus.Logger) error {
	level, err := c.LogLevel()
	if err != nil {
		return err
	}

	logger.SetLevel(level)
	return nil
}

func (c *ModelineConfig) BindFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("parents", "p", false, "create parent directories if they don't exist")
	c.BindPFlag("parents", cmd.Flags().Lookup("parents"))
	c.BindEnv("parents")

	cmd.Flags().String("version", "", "chart version (for remote charts)")
	c.BindPFlag("version", cmd.Flags().Lookup("version"))
	c.BindEnv("version")

	cmd.Flags().String("log-level", "warn", "log level (debug, info, warn, error, fatal, panic)")
	c.BindPFlag("log-level", cmd.Flags().Lookup("log-level"))
	c.BindEnv("log-level")
}

func (c *ModelineConfig) ToPackageConfig(chartRef string, targetFile string) *modeline.Config {
	return &modeline.Config{
		ChartRef:        chartRef,
		ChartVersion:    c.GetString("version"),
		TargetFile:      targetFile,
		CreateParents:   c.GetBool("parents"),
		PartialModeline: modeline.NewPartialModeline("yaml-language-server", "$schema"),
	}
}
