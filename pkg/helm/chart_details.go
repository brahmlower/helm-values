package helm

import (
	"fmt"
	"helmvalues/pkg/charts"
	"path/filepath"
)

func ChartDetailsFromRef(chartRef *ChartRef) (*charts.ChartDetails, error) {
	if chartRef.Path != "" {
		return ChartDetailsFromPath(chartRef.Path)
	}

	return ChartDetailsFromCache(chartRef)
}

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
