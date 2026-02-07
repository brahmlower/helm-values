package schema

import "github.com/sirupsen/logrus"

type Config struct {
	StdOut        bool
	Strict        bool
	DryRun        bool
	GitAdd        bool
	WriteModeline bool
	LogLevel      logrus.Level
}
