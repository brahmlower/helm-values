package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSchemaConfig() *SchemaConfig {
	cfg := standardViper()

	return &SchemaConfig{cfg}
}

type SchemaConfig struct {
	*viper.Viper
}

func (c *SchemaConfig) SchemaFile() string {
	return c.GetString("schema-file")
}

func (c *SchemaConfig) StdOut() bool {
	return c.GetBool("stdout")
}

func (c *SchemaConfig) Strict() bool {
	return c.GetBool("strict")
}

func (c *SchemaConfig) DryRun() bool {
	return c.GetBool("dry-run")
}

func (c *SchemaConfig) WriteModeline() bool {
	return c.GetBool("write-modeline")
}

func (c *SchemaConfig) LogLevel() (logrus.Level, error) {
	return logrus.ParseLevel(c.GetString("log-level"))
}

func (c *SchemaConfig) UpdateLogger(logger *logrus.Logger) error {
	level, err := c.LogLevel()
	if err != nil {
		return err
	}

	logger.SetLevel(level)
	return nil
}

func (c *SchemaConfig) BindFlags(cmd *cobra.Command) {
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

	cmd.Flags().Bool("write-modeline", true, "write modeline to values file")
	c.BindPFlag("write-modeline", cmd.Flags().Lookup("write-modeline"))
	c.BindEnv("write-modeline")
}
