package helm

import (
	"fmt"
	"strings"
)

type ChartRef struct {
	Repository string
	Chart      string
	Version    string
	Path       string
}

func NewChartRef(ref string) (*ChartRef, error) {
	parts1 := strings.SplitN(ref, "/", 2)
	if len(parts1) != 2 {
		cr := &ChartRef{
			Path: ref,
		}
		return cr, nil
	}
	repository := parts1[0]

	parts2 := strings.SplitN(parts1[1], "@", 2)
	chart := parts2[0]
	version := ""
	if len(parts2) == 2 {
		version = parts2[1]
	}

	cr := &ChartRef{
		Repository: repository,
		Chart:      chart,
		Version:    version,
	}
	return cr, nil
}

func (c ChartRef) String() string {
	if c.Version != "" {
		return fmt.Sprintf("%s/%s@%s", c.Repository, c.Chart, c.Version)
	}
	return fmt.Sprintf("%s/%s", c.Repository, c.Chart)
}
