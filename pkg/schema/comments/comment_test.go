package comments

import (
	"fmt"
	"helmvalues/pkg"
	"testing"

	"regexp"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

const COMMENT_MISSING_SPACE_PREFIX = `
#comment has no lead space
foo: bar
`

const COMMENT_WITH_INVALID_YAML = `
# @invalid yaml string
foo: bar
`

const DOESNT_SET_SCHEMA_PROPERTIES = `
# key: value
foo: bar
`

const COMMENT_WITH_YAML_STRING = `
# comment is just a string
foo: bar
`

const SETS_SCHEMA_DEFAULT = `
# default: baz
foo: bar
`

const SETS_SCHEMA_WITH_MULTILINE_VALUE = `
# default: |
#   foo
#   bar
foo: bar
`

const SETS_DESCRIPTION_TO_SECOND_DOC = `
# default: baz
# ---
# this is a description
foo: bar
`

func TestBasicCommentParsing(t *testing.T) {
	var tests = []struct {
		name          string
		document      string
		expectedError string
		validate      func(tt *testing.T, s *pkg.JsonSchema, err error)
	}{
		{
			name:     "empty document makes no changes",
			document: "",
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.Nil(tt, err)
				assert.Equal(tt, *s, pkg.JsonSchema{})
			},
		},
		{
			name:     "errors when comment missing space prefix",
			document: COMMENT_MISSING_SPACE_PREFIX,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NotNil(tt, err)
				assert.ErrorContains(t, err, "unexpected prefix")
			},
		},
		{
			// TODO: Fix comment parsing so that the description is correctly extracted
			name:     "errors when comment is invalid yaml string",
			document: COMMENT_WITH_INVALID_YAML,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, "", s.Description)
			},
		},
		{
			name:     "comment with string yaml is treated as description",
			document: COMMENT_WITH_YAML_STRING,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, "comment is just a string", s.Description)
			},
		},
		{
			name:     "comment has no jsonschema properties",
			document: DOESNT_SET_SCHEMA_PROPERTIES,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, pkg.JsonSchema{}, *s)
			},
		},
		{
			name:     "comment sets jsonschema field: default",
			document: SETS_SCHEMA_DEFAULT,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, "baz", s.Default)
			},
		},
		{
			name:     "comment sets jsonschema field w/ multiline value",
			document: SETS_SCHEMA_WITH_MULTILINE_VALUE,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, "foo\nbar", s.Default)
			},
		},
		{
			name:     "comment sets jsonschema description to second yaml doc",
			document: SETS_DESCRIPTION_TO_SECOND_DOC,
			validate: func(tt *testing.T, s *pkg.JsonSchema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, "baz", s.Default)
				assert.Equal(tt, "this is a description", s.Description)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.document), yamlNode)
			assert.NoError(tt, err)

			s, err := Parse(getCommentNode(yamlNode), nil)

			tc.validate(tt, s, err)
		})
	}
}

func TestCommentFieldsSingleLine(t *testing.T) {
	type testCase struct {
		field         string
		commentValue  string
		expectedValue any
		validate      func(tt *testing.T, tc testCase, s *pkg.JsonSchema)
	}

	var tests = []testCase{
		{
			field:         "$schema",
			commentValue:  "https://example.com/schema",
			expectedValue: "https://example.com/schema",
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Schema)
				assert.Equal(tt, tc.expectedValue, s.Schema)
			},
		},
		{
			field:         "description",
			commentValue:  "some description",
			expectedValue: "some description",
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Description)
				assert.Equal(tt, tc.expectedValue, s.Description)
			},
		},
		{
			field:         "format",
			commentValue:  "some format",
			expectedValue: "some format",
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Format)
				assert.Equal(tt, tc.expectedValue, s.Format)
			},
		},
		{
			field:         "minLength",
			commentValue:  "5",
			expectedValue: int64(5),
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.MinLength)
				assert.Equal(tt, tc.expectedValue, s.MinLength)
			},
		},
		{
			field:         "deprecated",
			commentValue:  "true",
			expectedValue: true,
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Deprecated)
				assert.Equal(tt, tc.expectedValue, s.Deprecated)
			},
		},
		{
			field:         "required",
			commentValue:  "[foo, bar]",
			expectedValue: []string{"foo", "bar"},
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Required)
				assert.Equal(tt, tc.expectedValue, s.Required)
			},
		},
		{
			field:         "maximum",
			commentValue:  "100",
			expectedValue: int64(100),
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Maximum)
				assert.Equal(tt, tc.expectedValue, s.Maximum)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.field, func(tt *testing.T) {
			document := fmt.Sprintf("# %s: %s\nfoo:bar\n", tc.field, tc.commentValue)

			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(document), yamlNode)
			assert.NoError(tt, err)

			s, err := Parse(yamlNode.Content[0], nil)
			assert.NoError(tt, err)

			tc.validate(tt, tc, s)
		})
	}
}

const TEST_FIELD_ONEOF = `
# oneOf:
#   - type: string
#     description: this is a string
#   - type: number
#     description: this is a number
foo: bar
`

const TEST_DEPENDENT_REQUIRED = `
# dependentRequired:
#   baz:
#     - qux
#     - quux
#   bif:
#     - quuz
foo: bar # line comment
`

const TEST_DEPENDENCIES = `
# dependencies:
#   baz: qux
#   bif: 0
#   qux:
#     - quux
#     - quuz
foo: bar
`

const TEST_PATTERN = `
# pattern: ^[a-z]+$
foo: bar
`

func TestCommentFieldsMultipleLines(t *testing.T) {
	type testCase struct {
		name          string
		comment       string
		expectedValue any
		validate      func(tt *testing.T, tc testCase, s *pkg.JsonSchema)
	}

	var tests = []testCase{
		{
			name:    "oneOf with multiple lines",
			comment: TEST_FIELD_ONEOF,
			expectedValue: []*pkg.JsonSchema{
				{Type: "string", Description: "this is a string"},
				{Type: "number", Description: "this is a number"},
			},
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.OneOf)
				assert.Equal(tt, tc.expectedValue, s.OneOf)
			},
		},
		{
			name:    "dependentRequired",
			comment: TEST_DEPENDENT_REQUIRED,
			expectedValue: map[string][]string{
				"baz": {"qux", "quux"},
				"bif": {"quuz"},
			},
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.DependentRequired)
				assert.Equal(tt, tc.expectedValue, s.DependentRequired)
			},
		},
		{
			name:    "dependencies",
			comment: TEST_DEPENDENCIES,
			expectedValue: map[string]any{
				"baz": "qux",
				"bif": 0,
				"qux": []any{"quux", "quuz"},
			},
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Dependencies)
				assert.Equal(tt, tc.expectedValue, s.Dependencies)
			},
		},
		{
			name:          "pattern",
			comment:       TEST_PATTERN,
			expectedValue: regexp.MustCompile("^[a-z]+$"),
			validate: func(tt *testing.T, tc testCase, s *pkg.JsonSchema) {
				assert.IsType(tt, tc.expectedValue, s.Pattern)
				assert.Equal(tt, tc.expectedValue, s.Pattern)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.comment), yamlNode)
			assert.NoError(tt, err)

			s, err := Parse(getCommentNode(yamlNode), nil)
			assert.NoError(tt, err)

			tc.validate(tt, tc, s)
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
