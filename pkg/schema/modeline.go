package schema

import (
	"fmt"
	"helmvalues/pkg/charts"
	"helmvalues/pkg/modeline"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func NewModelineWriter(filepath string) (*ModelineWriter, error) {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ml := &ModelineWriter{
		filepath: filepath,
		content:  string(content),
	}

	return ml, nil
}

type ModelineWriter struct {
	filepath string
	content  string
}

func (w *ModelineWriter) SetModeline(value *modeline.Modeline) {
	w.content = value.String() + "\n" + w.content
}

func (w *ModelineWriter) WriteToFile() error {
	return os.WriteFile(w.filepath, []byte(w.content), 0644)
}

func WriteSchemaModeline(logger *logrus.Logger, chart *charts.Chart, valuesPath string, dryRun bool) error {
	mlWriter, err := NewModelineWriter(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to read values file: %w", err)
	}

	ml := modeline.NewModeline("yaml-language-server", "$schema", chart.SchemaFilePath())
	mlWriter.SetModeline(ml)

	if dryRun {
		logger.Infof("schema: %s: dry-run enabled, skipping modeline write to %s", chart.Details.Name, valuesPath)
		return nil
	}

	return mlWriter.WriteToFile()
}
