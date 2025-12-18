package commands

import (
	"strings"

	"github.com/spf13/viper"
)

func standardViper() *viper.Viper {
	cfg := viper.New()
	cfg.AllowEmptyEnv(true)
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return cfg
}
