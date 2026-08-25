// Package schema generates a JSON Schema from a chart's values file and
// writes it to disk.
package schema

import "github.com/sirupsen/logrus"

// Config controls how a chart's values schema is generated and written.
type Config struct {
	StdOut        bool
	Strict        bool
	DryRun        bool
	GitAdd        bool
	WriteModeline bool
	Check         bool
	LogLevel      logrus.Level
}
