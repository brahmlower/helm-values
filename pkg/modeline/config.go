package modeline

import "helmvalues/pkg/helm"

// Config holds configuration for the modeline command
type Config struct {
	ChartRef        *helm.ChartRef
	TargetFile      string
	CreateParents   bool
	PartialModeline PartialModeline
}
