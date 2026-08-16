// Package modeline reads and writes the values-schema modeline comment
// (a "program: key=value" line) inside a chart's values file.
package modeline

import "helmvalues/pkg/helm"

// Config holds configuration for the modeline command.
type Config struct {
	ChartRef        *helm.ChartRef
	TargetFile      string
	CreateParents   bool
	PartialModeline PartialModeline
}
