// Package config binds cobra command flags and environment variables to
// viper, and translates the bound values into the config types each
// helm-values subcommand's underlying package expects.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logLevelFlag is the flag/env name shared by every subcommand's log-level
// setting.
const logLevelFlag = "log-level"

func standardViper() *viper.Viper {
	cfg := viper.New()
	cfg.AllowEmptyEnv(true)
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return cfg
}

// bindFlag binds the cobra flag named name (already registered on cmd) to
// v, along with its environment-variable equivalent.
func bindFlag(v *viper.Viper, cmd *cobra.Command, name string) error {
	if err := v.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
		return fmt.Errorf("binding %s flag: %w", name, err)
	}

	if err := v.BindEnv(name); err != nil {
		return fmt.Errorf("binding %s env: %w", name, err)
	}

	return nil
}
