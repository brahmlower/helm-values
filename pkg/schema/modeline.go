package schema

import (
	"fmt"
	"helmvalues/pkg/charts"
	"helmvalues/pkg/modeline"

	"github.com/sirupsen/logrus"
)

func WriteSchemaModeline(logger *logrus.Logger, chart *charts.Chart, valuesPath string, dryRun bool) error {
	fileManager, err := modeline.NewFileModelineManager(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to read values file: %w", err)
	}

	ml := modeline.NewModeline("yaml-language-server", "$schema", chart.SchemaFilePath())
	fileManager.SetModeline(ml)

	if dryRun {
		logger.Infof("schema: %s: dry-run enabled, skipping modeline write to %s", chart.Details.Name, valuesPath)
		return nil
	}

	return fileManager.Write(false)
}
