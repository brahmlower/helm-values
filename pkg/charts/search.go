package charts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Search walks the given chart directories (supporting glob patterns) and
// returns all Helm charts found within them.
func Search(logger *logrus.Logger, chartDirs []string) ([]*Chart, error) {
	cleanedChartDirs, err := cleanPaths(chartDirs)
	if err != nil {
		return nil, err
	}

	foundCharts := []*Chart{}

	for _, rootDir := range cleanedChartDirs {
		err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			logger.Tracef("search: checking path: %s", path)

			chart := findChartAtPath(logger, path)
			if chart != nil {
				logger.Infof("search: found chart %s at %s", chart.Details.Name, path)
				foundCharts = append(foundCharts, chart)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk chart directory %q: %w", rootDir, err)
		}
	}

	logger.Debugf("search: found %d charts", len(foundCharts))

	return foundCharts, nil
}

// findChartAtPath checks whether path contains a Helm chart (i.e. a
// Chart.yaml and values.yaml file) and, if so, loads and returns it. It
// returns nil if path does not contain a chart, or the chart fails to load.
func findChartAtPath(logger *logrus.Logger, path string) *Chart {
	chartFileInfo, err := os.Stat(path + "/Chart.yaml")
	if err != nil {
		logger.
			WithField("reason", "error").
			WithError(err).
			Tracef("search: skipping path: %s", path)

		return nil
	}

	if chartFileInfo.IsDir() {
		logger.
			WithField("reason", "Chart.yaml is a directory").
			Tracef("search: skipping path: %s", path)

		return nil
	}

	valuesFileInfo, err := os.Stat(path + "/Chart.yaml")
	if err != nil {
		logger.
			WithField("reason", "error").
			WithError(err).
			Tracef("search: kipping path: %s", path)

		return nil
	}

	if valuesFileInfo.IsDir() {
		logger.
			WithField("reason", "values.yaml is a directory").
			Tracef("search: skipping path: %s", path)

		return nil
	}

	chart, err := NewChart(path)
	if err != nil {
		logger.
			WithField("reason", "error").
			WithError(err).
			Warnf("search: skipping possible chart: %s", path)

		return nil
	}

	return chart
}

func cleanPaths(paths []string) ([]string, error) {
	cleanedPaths := []string{}

	for _, path := range paths {
		ap, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve absolute path %q: %w", path, err)
		}

		ps, err := filepath.Glob(ap)
		if err != nil {
			return nil, fmt.Errorf("expand glob pattern %q: %w", ap, err)
		}

		cleanedPaths = append(cleanedPaths, ps...)
	}

	return cleanedPaths, nil
}
