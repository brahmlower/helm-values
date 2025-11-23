package charts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func Search(logger *logrus.Logger, chartDirs []string) ([]*Chart, error) {
	foundCharts := []*Chart{}
	for _, rootDir := range chartDirs {
		err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}

			logger.Tracef("Checking path: %s", path)

			chartFileInfo, err := os.Stat(fmt.Sprintf("%s/Chart.yaml", path))
			if err != nil {
				logger.
					WithField("reason", "error").
					WithError(err).
					Tracef("Skipping path: %s", path)
				return nil
			}
			if chartFileInfo.IsDir() {
				logger.
					WithField("reason", "Chart.yaml is a directory").
					Tracef("Skipping path: %s", path)
				return nil
			}

			valuesFileInfo, err := os.Stat(fmt.Sprintf("%s/Chart.yaml", path))
			if err != nil {
				logger.
					WithField("reason", "error").
					WithError(err).
					Tracef("Skipping path: %s", path)
				return nil
			}
			if valuesFileInfo.IsDir() {
				logger.
					WithField("reason", "values.yaml is a directory").
					Tracef("Skipping path: %s", path)
				return nil
			}

			logger.Infof("Found possible chart: %s", path)

			chart, err := NewChart(path)
			if err != nil {
				logger.
					WithField("reason", "error").
					WithError(err).
					Warnf("Skipping possible chart: %s", path)
				return nil
			}

			foundCharts = append(foundCharts, chart)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return foundCharts, nil
}
