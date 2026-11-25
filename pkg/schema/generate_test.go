package schema_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func TestGenerateSetsGeneratedComment(t *testing.T) {
	t.Parallel()

	s, err := generate(t, "# a value\nfoo: 1\n", false)
	require.NoError(t, err)

	assert.Equal(t, schema.GeneratedComment, s.Comment)

	content, err := json.MarshalIndent(s, "", "  ")
	require.NoError(t, err)

	// "$comment" is the first key, so it is the first thing a reader sees
	assert.True(t, strings.HasPrefix(string(content), "{\n  \"$comment\": "), "got:\n%s", content)

	// nested properties don't get the comment
	assert.Empty(t, property(t, s, "foo").Comment)
}

func TestGenerateKeepsPropertyComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values string
		want   string
	}{
		{
			name:   "$comment",
			values: "# $comment: maintainer note\nfoo: 1\n",
			want:   "maintainer note",
		},
		{
			name:   "legacy comment key",
			values: "# comment: old note\nfoo: 1\n",
			want:   "old note",
		},
		{
			name:   "$comment wins over legacy comment key",
			values: "# comment: old note\n# $comment: new note\nfoo: 1\n",
			want:   "new note",
		},
		{
			name:   "$comment wins regardless of order",
			values: "# $comment: new note\n# comment: old note\nfoo: 1\n",
			want:   "new note",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, err := generate(t, tc.values, true)
			require.NoError(t, err)

			assert.Equal(t, schema.GeneratedComment, s.Comment)
			assert.Equal(t, tc.want, property(t, s, "foo").Comment)
		})
	}
}

// TestGenerateIgnoresModelineInFirstKeyComment is the issue #39 parser case:
// a modeline directly above the first key's doc comment.
func TestGenerateIgnoresModelineInFirstKeyComment(t *testing.T) {
	t.Parallel()

	values := "# yaml-language-server: $schema=values.schema.json\n# -- replicas to run\nreplicas: 1\n"

	s, err := generate(t, values, false)
	require.NoError(t, err)

	assert.Equal(t, "-- replicas to run", property(t, s, "replicas").Description)
}

// TestGenerateSchemaThenCheckWithModeline is the issue #39 reproduction: a
// fresh generate that writes the modeline must leave --check passing.
func TestGenerateSchemaThenCheckWithModeline(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	valuesPath := filepath.Join(dir, "values.yaml")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(testChartFile), 0o600))
	require.NoError(t, os.WriteFile(valuesPath, []byte("# -- replicas to run\nreplicas: 1\n"), 0o600))

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	cfg := &schema.Config{
		StdOut:        false,
		Strict:        false,
		DryRun:        false,
		GitAdd:        false,
		WriteModeline: true,
		Check:         false,
		LogLevel:      logrus.InfoLevel,
	}
	require.NoError(t, schema.GenerateSchema(logger, cfg, []string{dir}))

	cfg.Check = true
	require.NoError(t, schema.GenerateSchema(logger, cfg, []string{dir}))

	values, err := fs.ReadFile(os.DirFS(dir), "values.yaml")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(values), "# yaml-language-server: $schema=values.schema.json\n\n"), "got:\n%s", values)

	schemaContent, err := fs.ReadFile(os.DirFS(dir), "values.schema.json")
	require.NoError(t, err)
	assert.Contains(t, string(schemaContent), "-- replicas to run")
}

func marshalProperties(t *testing.T, s *pkg.JsonSchema) map[string]json.RawMessage {
	t.Helper()

	content, err := json.Marshal(s)
	require.NoError(t, err)

	var out struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(content, &out))

	return out.Properties
}

// TestGenerateConst is the issue #41 reproduction: "const" is written under
// its draft-07 keyword, including null and inside nested schemas.
func TestGenerateConst(t *testing.T) {
	t.Parallel()

	values := "# const: a\nscalar: a\n" +
		"# const: [a, b]\nlist: []\n" +
		"# const: {a: 1}\nobject: {}\n" +
		"# const: null\nnull: null\n" +
		"# oneOf: [{const: null}, {type: string}]\nnested: null\n"

	s, err := generate(t, values, true)
	require.NoError(t, err)

	props := marshalProperties(t, s)

	assert.Contains(t, string(props["scalar"]), `"const":"a"`)
	assert.Contains(t, string(props["list"]), `"const":["a","b"]`)
	assert.Contains(t, string(props["object"]), `"const":{"a":1}`)
	assert.Contains(t, string(props["null"]), `"const":null`)
	assert.Contains(t, string(props["nested"]), `"oneOf":[{"const":null},{"type":"string"}]`)
}

// TestGenerateDropsNonStandardKeys pins down that the old non-standard keys
// (including the legacy "constant") are ignored, not aliased, and aren't a
// strict-mode error (issue #41).
func TestGenerateDropsNonStandardKeys(t *testing.T) {
	t.Parallel()

	values := "# constant: [a, b]\n# recursiveAnchor: true\n# recursiveRef: \"#\"\n" +
		"# dynamicAnchor: x\n# dynamicRef: \"#x\"\n# always: true\n# location: here\nfoo: a\n" +
		"# constant: foo\n# ---\n# scalar constant\nbar: a\n"

	s, err := generate(t, values, true)
	require.NoError(t, err)

	props := marshalProperties(t, s)

	for _, key := range []string{"constant", "recursiveAnchor", "recursiveRef", "dynamicAnchor", "dynamicRef", "always", "location", "const"} {
		for _, prop := range []string{"foo", "bar"} {
			assert.NotContains(t, string(props[prop]), `"`+key+`"`, "%s has %s", prop, key)
		}
	}

	assert.Equal(t, "scalar constant", property(t, s, "bar").Description)
}
