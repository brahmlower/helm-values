package comments_test

import (
	"fmt"
	"testing"

	"helmvalues/pkg"
	"helmvalues/pkg/schema/comments"

	"regexp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

const CommentMissingSpacePrefix = `
#comment has no lead space
foo: bar
`

const CommentWithInvalidYAML = `
# @invalid yaml string
foo: bar
`

const DoesntSetSchemaProperties = `
# key: value
foo: bar
`

const CommentWithYAMLString = `
# comment is just a string
foo: bar
`

const CommentWithColonAndMultiWordKey = `
# See docs for details: https://example.com
foo: bar
`

const CommentWithColonAndTwoWordKey = `
# Contact us: support@example.com
foo: bar
`

const SetsSchemaDefault = `
# default: baz
foo: bar
`

const SetsSchemaWithMultilineValue = `
# default: |
#   foo
#   bar
foo: bar
`

const SetsDescriptionToSecondDoc = `
# default: baz
# ---
# this is a description
foo: bar
`

// testQux is the shared "qux" value referenced across the dependentRequired
// and dependencies test cases below.
const testQux = "qux"

func TestBasicCommentParsing(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name     string
		document string
		validate func(t *testing.T, s *pkg.JsonSchema, err error)
	}{
		{
			name:     "empty document makes no changes",
			document: "",
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, *s)
			},
		},
		{
			name:     "errors when comment missing space prefix",
			document: CommentMissingSpacePrefix,
			validate: func(t *testing.T, _ *pkg.JsonSchema, err error) {
				t.Helper()
				require.Error(t, err)
				assert.ErrorContains(t, err, "unexpected prefix")
			},
		},
		{
			// TODO: Fix comment parsing so that the description is correctly extracted
			name:     "errors when comment is invalid yaml string",
			document: CommentWithInvalidYAML,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, s.Description)
			},
		},
		{
			name:     "comment with string yaml is treated as description",
			document: CommentWithYAMLString,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "comment is just a string", s.Description)
			},
		},
		{
			name:     "comment with a colon after a multi-word phrase is treated as description",
			document: CommentWithColonAndMultiWordKey,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "See docs for details: https://example.com", s.Description)
			},
		},
		{
			name:     "comment with a colon after a two-word phrase is treated as description",
			document: CommentWithColonAndTwoWordKey,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "Contact us: support@example.com", s.Description)
			},
		},
		{
			name:     "comment has no jsonschema properties",
			document: DoesntSetSchemaProperties,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, *s)
			},
		},
		{
			name:     "comment sets jsonschema field: default",
			document: SetsSchemaDefault,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "baz", s.Default)
			},
		},
		{
			name:     "comment sets jsonschema field w/ multiline value",
			document: SetsSchemaWithMultilineValue,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "foo\nbar", s.Default)
			},
		},
		{
			name:     "comment sets jsonschema description to second yaml doc",
			document: SetsDescriptionToSecondDoc,
			validate: func(t *testing.T, s *pkg.JsonSchema, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "baz", s.Default)
				assert.Equal(t, "this is a description", s.Description)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.document), yamlNode)
			require.NoError(t, err)

			s, err := comments.Parse(getCommentNode(yamlNode), nil)

			tc.validate(t, s, err)
		})
	}
}

// TestParseOverridesExtraNodeWithCommentField exercises the scenario that
// used to panic: extraNodes (auto-derived fields, e.g. an inferred "type")
// and a comment field share the same key. The comment-authored value should
// win instead of producing a duplicate-mapping-key error.
func TestParseOverridesExtraNodeWithCommentField(t *testing.T) {
	t.Parallel()

	extraNodes := comments.KeyValueNodes("type", "object")

	document := "# type: string\nfoo: bar\n"

	yamlNode := &yaml.Node{}
	err := yaml.Unmarshal([]byte(document), yamlNode)
	require.NoError(t, err)

	s, err := comments.Parse(getCommentNode(yamlNode), extraNodes)
	require.NoError(t, err)
	assert.Equal(t, "string", s.Type)
}

// TestColonDescriptionWordCountBoundary sweeps the word count of the phrase
// before a colon to pin the exact boundary of the heuristic in
// commentAsDescriptionNodes: any phrase with 2+ words before the colon is
// rescued as a description, but a single word is indistinguishable from an
// unrecognized schema keyword (like "key: value" in DoesntSetSchemaProperties
// above) and is silently dropped. That single-word case is a known,
// intentional-but-surprising limitation, not something this test expects to
// change without a deliberate decision.
func TestColonDescriptionWordCountBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		phraseBeforeColon string
		value             string
		expectDescription bool
	}{
		{
			name:              "single word before colon is dropped (known limitation)",
			phraseBeforeColon: "Note",
			value:             "this is important",
			expectDescription: false,
		},
		{
			name:              "two words before colon is a description",
			phraseBeforeColon: "Contact us",
			value:             "support@example.com",
			expectDescription: true,
		},
		{
			name:              "three words before colon is a description",
			phraseBeforeColon: "See docs for",
			value:             "https://example.com",
			expectDescription: true,
		},
		{
			name:              "five words before colon is a description",
			phraseBeforeColon: "See docs for full details",
			value:             "https://example.com",
			expectDescription: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expected := tc.phraseBeforeColon + ": " + tc.value
			document := "# " + expected + "\nfoo: bar\n"

			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(document), yamlNode)
			require.NoError(t, err)

			s, err := comments.Parse(getCommentNode(yamlNode), nil)
			require.NoError(t, err)

			if tc.expectDescription {
				assert.Equal(t, expected, s.Description)
			} else {
				assert.Empty(t, *s)
			}
		})
	}
}

// TestColonDescriptionAfterDelimiter confirms that a colon-containing
// description is rescued the same way whether it appears before or after
// the "---" doc separator: parseNodeComment splits the head comment on
// "---" into independent segments before commentAsDescriptionNodes runs on
// each one, so the fix for the boundary above applies regardless of
// position.
func TestColonDescriptionAfterDelimiter(t *testing.T) {
	t.Parallel()

	document := `
# default: baz
# ---
# Contact us: support@example.com
foo: bar
`

	yamlNode := &yaml.Node{}
	err := yaml.Unmarshal([]byte(document), yamlNode)
	require.NoError(t, err)

	s, err := comments.Parse(getCommentNode(yamlNode), nil)
	require.NoError(t, err)
	assert.Equal(t, "baz", s.Default)
	assert.Equal(t, "Contact us: support@example.com", s.Description)
}

func TestCommentFieldsSingleLine(t *testing.T) {
	t.Parallel()

	type testCase struct {
		field         string
		commentValue  string
		expectedValue any
		validate      func(t *testing.T, tc testCase, s *pkg.JsonSchema)
	}

	var tests = []testCase{
		{
			field:         "$schema",
			commentValue:  "https://example.com/schema",
			expectedValue: "https://example.com/schema",
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Schema)
				assert.Equal(t, tc.expectedValue, s.Schema)
			},
		},
		{
			field:         "description",
			commentValue:  "some description",
			expectedValue: "some description",
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Description)
				assert.Equal(t, tc.expectedValue, s.Description)
			},
		},
		{
			field:         "format",
			commentValue:  "some format",
			expectedValue: "some format",
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Format)
				assert.Equal(t, tc.expectedValue, s.Format)
			},
		},
		{
			field:         "minLength",
			commentValue:  "5",
			expectedValue: int64(5),
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.MinLength)
				assert.Equal(t, tc.expectedValue, s.MinLength)
			},
		},
		{
			field:         "deprecated",
			commentValue:  "true",
			expectedValue: true,
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Deprecated)
				assert.Equal(t, tc.expectedValue, s.Deprecated)
			},
		},
		{
			field:         "required",
			commentValue:  "[foo, bar]",
			expectedValue: []string{"foo", "bar"},
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Required)
				assert.Equal(t, tc.expectedValue, s.Required)
			},
		},
		{
			field:         "maximum",
			commentValue:  "100",
			expectedValue: int64(100),
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Maximum)
				assert.Equal(t, tc.expectedValue, s.Maximum)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			t.Parallel()

			document := fmt.Sprintf("# %s: %s\nfoo:bar\n", tc.field, tc.commentValue)

			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(document), yamlNode)
			require.NoError(t, err)

			s, err := comments.Parse(yamlNode.Content[0], nil)
			require.NoError(t, err)

			tc.validate(t, tc, s)
		})
	}
}

const TestFieldOneOf = `
# oneOf:
#   - type: string
#     description: this is a string
#   - type: number
#     description: this is a number
foo: bar
`

const TestDependentRequired = `
# dependentRequired:
#   baz:
#     - qux
#     - quux
#   bif:
#     - quuz
foo: bar # line comment
`

const TestDependencies = `
# dependencies:
#   baz: qux
#   bif: 0
#   qux:
#     - quux
#     - quuz
foo: bar
`

const TestPattern = `
# pattern: ^[a-z]+$
foo: bar
`

func TestCommentFieldsMultipleLines(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		comment       string
		expectedValue any
		validate      func(t *testing.T, tc testCase, s *pkg.JsonSchema)
	}

	oneOfStringSchema := pkg.NewJsonSchema()
	oneOfStringSchema.Type = "string"
	oneOfStringSchema.Description = "this is a string"

	oneOfNumberSchema := pkg.NewJsonSchema()
	oneOfNumberSchema.Type = "number"
	oneOfNumberSchema.Description = "this is a number"

	var tests = []testCase{
		{
			name:          "oneOf with multiple lines",
			comment:       TestFieldOneOf,
			expectedValue: []*pkg.JsonSchema{oneOfStringSchema, oneOfNumberSchema},
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.OneOf)
				assert.Equal(t, tc.expectedValue, s.OneOf)
			},
		},
		{
			name:    "dependentRequired",
			comment: TestDependentRequired,
			expectedValue: map[string][]string{
				"baz": {testQux, "quux"},
				"bif": {"quuz"},
			},
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.DependentRequired)
				assert.Equal(t, tc.expectedValue, s.DependentRequired)
			},
		},
		{
			name:    "dependencies",
			comment: TestDependencies,
			expectedValue: map[string]any{
				"baz": testQux,
				"bif": 0,
				"qux": []any{"quux", "quuz"},
			},
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Dependencies)
				assert.Equal(t, tc.expectedValue, s.Dependencies)
			},
		},
		{
			name:          "pattern",
			comment:       TestPattern,
			expectedValue: regexp.MustCompile("^[a-z]+$"),
			validate: func(t *testing.T, tc testCase, s *pkg.JsonSchema) {
				t.Helper()
				assert.IsType(t, tc.expectedValue, s.Pattern)
				assert.Equal(t, tc.expectedValue, s.Pattern)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.comment), yamlNode)
			require.NoError(t, err)

			s, err := comments.Parse(getCommentNode(yamlNode), nil)
			require.NoError(t, err)

			tc.validate(t, tc, s)
		})
	}
}

func getCommentNode(node *yaml.Node) *yaml.Node {
	if len(node.Content) == 0 {
		return node
	}

	// Get the first scalar node in the document
	return node.Content[0].Content[0]
}
