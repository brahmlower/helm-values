package docs_test

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"helmvalues/pkg/docs"
	"helmvalues/pkg/docs/templates"

	"github.com/samber/mo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	badCommentValues = "# minimum: .inf\nbad: 2\n"
	readmeFile       = "README.md"
)

// writeNamedChart writes a chart named name with the given values file.
func writeNamedChart(t *testing.T, name, values string) string {
	t.Helper()

	chartFile := fmt.Sprintf("apiVersion: v2\nname: %s\nversion: 0.1.0\ndescription: A test chart\n", name)

	return writeChart(t, map[string]string{"Chart.yaml": chartFile, "values.yaml": values})
}

func strictDocsConfig(strict bool) *docs.Config {
	cfg := docsConfig(mo.None[templates.Markup]())
	cfg.Strict = strict

	return cfg
}

func quietLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	return logger
}

func TestGenerateDocsStrictBadCommentFails(t *testing.T) {
	t.Parallel()

	dir := writeNamedChart(t, "bad", badCommentValues)

	err := docs.GenerateDocs(quietLogger(), strictDocsConfig(true), []string{dir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docs: bad:")
	assert.Contains(t, err.Error(), "doc comment error")
	assert.Equal(t, 1, strings.Count(err.Error(), "docs: "))
	assert.NoFileExists(t, filepath.Join(dir, readmeFile))
}

func TestGenerateDocsNonStrictBadCommentWarns(t *testing.T) {
	t.Parallel()

	dir := writeNamedChart(t, "bad", badCommentValues)

	require.NoError(t, docs.GenerateDocs(quietLogger(), strictDocsConfig(false), []string{dir}))
	assert.FileExists(t, filepath.Join(dir, readmeFile))
}

func TestGenerateDocsContinuesAfterFailingChart(t *testing.T) {
	t.Parallel()

	badDir := writeNamedChart(t, "bad", badCommentValues)
	goodDir := writeNamedChart(t, "good", testValuesFile)

	err := docs.GenerateDocs(quietLogger(), strictDocsConfig(true), []string{badDir, goodDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docs: bad:")
	assert.NotContains(t, err.Error(), "good")
	assert.FileExists(t, filepath.Join(goodDir, readmeFile))
	assert.NoFileExists(t, filepath.Join(badDir, readmeFile))
}
