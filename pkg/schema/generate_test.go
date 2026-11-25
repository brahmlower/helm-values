package schema_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"helmvalues/pkg"
	"helmvalues/pkg/charts"
	"helmvalues/pkg/schema"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testChartFile = "apiVersion: v2\nname: t\nversion: 0.1.0\n"

func generate(t *testing.T, values string, strict bool) (*pkg.JsonSchema, error) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(testChartFile), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(values), 0o600))

	chart, err := charts.NewChart(dir)
	require.NoError(t, err)

	cfg := &schema.Config{
		StdOut:        false,
		Strict:        strict,
		DryRun:        true,
		GitAdd:        false,
		WriteModeline: false,
		Check:         false,
		LogLevel:      logrus.InfoLevel,
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	s, err := schema.NewGenerator(logger, schema.NewPlan(cfg, chart)).Generate()
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	return s, nil
}

func property(t *testing.T, s *pkg.JsonSchema, key string) *pkg.JsonSchema {
	t.Helper()

	require.NotNil(t, s.Properties)

	prop, ok := s.Properties.Get(key)
	require.True(t, ok, "missing property %s", key)
	require.NotNil(t, prop, "nil schema for property %s", key)

	return prop
}

// TestGenerateKeepsZeroAndFractionalBounds is the issue #36 reproduction.
func TestGenerateKeepsZeroAndFractionalBounds(t *testing.T) {
	t.Parallel()

	values := "# maxLength: 0\nempty: \"\"\n# minimum: 0\nzero: 1\n# minimum: 0.5\nrate: 1.0\n"

	s, err := generate(t, values, false)
	require.NoError(t, err)

	content, err := json.Marshal(s)
	require.NoError(t, err)

	var out struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(content, &out))

	assert.Contains(t, out.Properties["empty"], "maxLength")
	assert.InDelta(t, 0.0, out.Properties["empty"]["maxLength"], 0)
	assert.Contains(t, out.Properties["zero"], "minimum")
	assert.InDelta(t, 0.0, out.Properties["zero"]["minimum"], 0)
	assert.InDelta(t, 0.5, out.Properties["rate"]["minimum"], 0)
}

func TestGenerateBadCommentNonStrict(t *testing.T) {
	t.Parallel()

	values := "# minimum: .inf\nbad: 1\n# a good value\ngood: 2\n"

	s, err := generate(t, values, false)
	require.NoError(t, err)

	bad := property(t, s, "bad")
	assert.Equal(t, "number", bad.Type)
	assert.Equal(t, "bad", bad.Title)
	assert.Equal(t, 1, bad.Default)
	assert.Nil(t, bad.Minimum)
	assert.Empty(t, bad.Description)

	good := property(t, s, "good")
	assert.Equal(t, "a good value", good.Description)
}

func TestGenerateBadCommentOnMappingNonStrict(t *testing.T) {
	t.Parallel()

	values := "# minimum: abc\nparent:\n  # child doc\n  child: 1\n"

	s, err := generate(t, values, false)
	require.NoError(t, err)

	parent := property(t, s, "parent")
	assert.Equal(t, "object", parent.Type)

	child := property(t, parent, "child")
	assert.Equal(t, "child doc", child.Description)
}

func TestGenerateBadCommentOnSequenceNonStrict(t *testing.T) {
	t.Parallel()

	values := "# minItems: abc\nlist:\n  - a\n"

	s, err := generate(t, values, false)
	require.NoError(t, err)

	list := property(t, s, "list")
	assert.Equal(t, "array", list.Type)
	assert.Nil(t, list.MinItems)
}

func TestGenerateBadCommentStrict(t *testing.T) {
	t.Parallel()

	values := "# minimum: .inf\nbad: 1\n# a good value\ngood: 2\n"

	_, err := generate(t, values, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "doc comment error")
	assert.Contains(t, err.Error(), "bad")
	assert.Contains(t, err.Error(), "not a finite number")
}

func TestGenerateNumberKeywordsAcceptYAMLForms(t *testing.T) {
	t.Parallel()

	values := "# minimum: .5\na: 1\n# maximum: 0x10\nb: 1\n# multipleOf: 1e-2\nc: 1\n"

	s, err := generate(t, values, true)
	require.NoError(t, err)

	assert.Equal(t, lo.ToPtr(pkg.Number(0.5)), property(t, s, "a").Minimum)
	assert.Equal(t, lo.ToPtr(pkg.Number(16)), property(t, s, "b").Maximum)
	assert.Equal(t, lo.ToPtr(pkg.Number(0.01)), property(t, s, "c").MultipleOf)
}
