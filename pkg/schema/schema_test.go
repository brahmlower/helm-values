package schema_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"helmvalues/pkg/schema"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	badCommentValues = "# minimum: .inf\nbad: 2\n"
	goodValues       = "# -- replicas to run\nreplicas: 1\n"
	schemaFile       = "values.schema.json"
)

// writeChartDir writes a chart named name with the given values file to its
// own temp dir and returns the dir.
func writeChartDir(t *testing.T, name, values string) string {
	t.Helper()

	dir := t.TempDir()
	chartFile := fmt.Sprintf("apiVersion: v2\nname: %s\nversion: 0.1.0\n", name)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartFile), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(values), 0o600))

	return dir
}

func schemaConfig(strict, check bool) *schema.Config {
	return &schema.Config{
		StdOut:        false,
		Strict:        strict,
		DryRun:        false,
		GitAdd:        false,
		WriteModeline: false,
		Check:         check,
		LogLevel:      logrus.InfoLevel,
	}
}

func quietLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	return logger
}

func TestGenerateSchemaStrictBadCommentFails(t *testing.T) {
	t.Parallel()

	dir := writeChartDir(t, "bad", badCommentValues)

	err := schema.GenerateSchema(quietLogger(), schemaConfig(true, false), []string{dir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema: bad:")
	assert.Contains(t, err.Error(), "doc comment error")
	assert.NoFileExists(t, filepath.Join(dir, schemaFile))
}

func TestGenerateSchemaNonStrictBadCommentWarns(t *testing.T) {
	t.Parallel()

	dir := writeChartDir(t, "bad", badCommentValues)

	require.NoError(t, schema.GenerateSchema(quietLogger(), schemaConfig(false, false), []string{dir}))
	assert.FileExists(t, filepath.Join(dir, schemaFile))
}

func TestGenerateSchemaContinuesAfterFailingChart(t *testing.T) {
	t.Parallel()

	badDir := writeChartDir(t, "bad", badCommentValues)
	goodDir := writeChartDir(t, "good", goodValues)

	err := schema.GenerateSchema(quietLogger(), schemaConfig(true, false), []string{badDir, goodDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema: bad:")
	assert.NotContains(t, err.Error(), "good")
	assert.FileExists(t, filepath.Join(goodDir, schemaFile))
	assert.NoFileExists(t, filepath.Join(badDir, schemaFile))
}

func TestGenerateSchemaJoinsMultipleChartErrors(t *testing.T) {
	t.Parallel()

	firstDir := writeChartDir(t, "first-bad", badCommentValues)
	secondDir := writeChartDir(t, "second-bad", badCommentValues)

	err := schema.GenerateSchema(quietLogger(), schemaConfig(true, false), []string{firstDir, secondDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema: first-bad:")
	assert.Contains(t, err.Error(), "schema: second-bad:")
}

func TestGenerateSchemaCheckReportsProblemsAlongsideErrors(t *testing.T) {
	t.Parallel()

	badDir := writeChartDir(t, "bad", badCommentValues)
	goodDir := writeChartDir(t, "good", goodValues)

	err := schema.GenerateSchema(quietLogger(), schemaConfig(true, true), []string{badDir, goodDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema: bad:")
	assert.Contains(t, err.Error(), "check failed: 1 problem(s) found")
}
