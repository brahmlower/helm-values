package helm

import (
	"fmt"
	"helmvalues/pkg/charts"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func ChartDetailsFromRef(chartRef string, version string) (*charts.ChartDetails, error) {
	refParts := strings.Split(chartRef, "/")
	if len(refParts) == 2 {
		return GetChartFromCache(logrus.StandardLogger(), chartRef, version)
	}

	return chartDetailsFromPath(chartRef)
}

func chartDetailsFromPath(chartRef string) (*charts.ChartDetails, error) {
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
