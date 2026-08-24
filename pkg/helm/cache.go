// Package helm provides helpers for working with Helm chart references,
// repository caches, and index files.
package helm

import (
	"fmt"
	"strings"
)

// maxSplitParts is the maximum number of parts to split a chart reference
// into when separating its repository/chart and chart@version segments.
const maxSplitParts = 2

// ChartRef represents a parsed reference to a Helm chart, either as a
// "repository/chart@version" reference or a local filesystem path.
type ChartRef struct {
	Repository string
	Chart      string
	Version    string
	Path       string
}

// NewChartRef parses a chart reference string into a ChartRef. Strings
// without a "/" are treated as local filesystem paths.
func NewChartRef(ref string) (*ChartRef, error) {
	parts1 := strings.SplitN(ref, "/", maxSplitParts)
	if len(parts1) != maxSplitParts {
		cr := &ChartRef{
			Repository: "",
			Chart:      "",
			Version:    "",
			Path:       ref,
		}

		return cr, nil
	}

	repository := parts1[0]

	parts2 := strings.SplitN(parts1[1], "@", maxSplitParts)
	chart := parts2[0]

	version := ""
	if len(parts2) == maxSplitParts {
		version = parts2[1]
	}

	cr := &ChartRef{
		Repository: repository,
		Chart:      chart,
		Version:    version,
		Path:       "",
	}

	return cr, nil
}

func (c ChartRef) String() string {
	if c.Version != "" {
		return fmt.Sprintf("%s/%s@%s", c.Repository, c.Chart, c.Version)
	}

	return fmt.Sprintf("%s/%s", c.Repository, c.Chart)
}
