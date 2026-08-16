package schema

import (
	"fmt"

	"helmvalues/pkg/charts"
	"helmvalues/pkg/modeline"

	"github.com/sirupsen/logrus"
)

// WriteSchemaModeline writes a yaml-language-server modeline pointing at the chart's
// generated schema file into the values file at valuesPath.
func WriteSchemaModeline(logger *logrus.Logger, chart *charts.Chart, valuesPath string, dryRun bool) error {
	fileManager, err := modeline.NewFileModelineManager(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to read values file: %w", err)
	}

	ml := modeline.NewModeline("yaml-language-server", "$schema", "values.schema.json")
	fileManager.SetModeline(ml)

	if dryRun {
		logger.Infof("schema: %s: dry-run enabled, skipping modeline write to %s", chart.Details.Name, valuesPath)

		return nil
	}

	if err := fileManager.Write(false); err != nil {
		return fmt.Errorf("failed to write modeline: %w", err)
	}

	return nil
}
