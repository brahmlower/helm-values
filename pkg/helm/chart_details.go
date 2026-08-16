package helm

import (
	"fmt"
	"path/filepath"

	"helmvalues/pkg/charts"
)

// ChartDetailsFromRef returns the chart details for the given ChartRef,
// loading them from a local path or, if no path is set, the repository cache.
func ChartDetailsFromRef(chartRef *ChartRef) (*charts.ChartDetails, error) {
	if chartRef.Path != "" {
		return ChartDetailsFromPath(chartRef.Path)
	}

	return ChartDetailsFromCache(chartRef)
}

// ChartDetailsFromCache returns the chart details for the given ChartRef by
// looking up its repository index in the local Helm repository cache.
func ChartDetailsFromCache(chartRef *ChartRef) (*charts.ChartDetails, error) {
	index, err := RepositoryIndexFromCache(chartRef.Repository)
	if err != nil {
		return nil, err
	}

	if chartRef.Version != "" {
		return index.GetVersion(chartRef.Chart, chartRef.Version)
	}

	return index.GetLatestVersion(chartRef.Chart)
}

// ChartDetailsFromPath loads chart details from the chart located at the
// given filesystem path.
func ChartDetailsFromPath(chartRef string) (*charts.ChartDetails, error) {
	absPath, err := filepath.Abs(chartRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve chart path: %w", err)
	}

	chart, err := charts.NewChart(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	return chart.Details, nil
}
