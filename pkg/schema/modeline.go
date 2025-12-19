package schema

import (
	"fmt"
	"helmvalues/internal/charts"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const YAML_MODELINE = "yaml-language-server"
const MODELINE_PARAM = "$schema="

func renderedModeline(schemaPath string) string {
	return fmt.Sprintf(
		"# %s: %s%s\n",
		YAML_MODELINE,
		MODELINE_PARAM,
		filepath.Base(schemaPath),
	)
}

func WriteSchemaModeline(logger *logrus.Logger, chart *charts.Chart, dryRun bool) error {
	valuesFilePath := chart.ValuesFilePath()

	if dryRun {
		logger.Infof("schema: %s: dry-run enabled, skipping modeline write to %s", chart.Details.Name, valuesFilePath)
		return nil
	}

	f, err := os.OpenFile(valuesFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	contentB, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	content := string(contentB)
	var updatedContent string

	modelineStart := strings.Index(content, fmt.Sprintf("# %s:", YAML_MODELINE))
	if modelineStart == -1 {
		// write an extra newline when inserting the modeline for the first time
		updatedContent = renderedModeline(chart.SchemaFilePath()) + "\n" + content
	} else {
		eolIdx := strings.Index(content[modelineStart:], "\n")
		updatedContent = content[:modelineStart] + renderedModeline(chart.SchemaFilePath()) + content[modelineStart+eolIdx+1:]
	}

	err = os.WriteFile(valuesFilePath, []byte(updatedContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
