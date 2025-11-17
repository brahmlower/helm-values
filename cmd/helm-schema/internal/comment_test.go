package internal

import (
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

const COMMENT_MISSING_SPACE_PREFIX = `
#comment has no lead space
foo: bar
`

const COMMENT_WITH_INVALID_YAML = `
# comment is not valid yaml
foo: bar
`

const DOESNT_SET_SCHEMA_PROPERTIES = `
# key: value
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

// const SETS_DESCRIPTION_TO_SECOND_DOC = `
// # default: baz
// # ---
// # this is a description
// foo: bar
// `

func TestNewClient(t *testing.T) {
	var tests = []struct {
		name          string
		document      string
		expectedError string
		validate      func(tt *testing.T, s *jsonschema.Schema, err error)
	}{
		{
			name:     "empty document makes no changes",
			document: "",
			validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
				assert.Nil(tt, err)
			},
		},
		{
			name:     "comment missing space prefix",
			document: COMMENT_MISSING_SPACE_PREFIX,
			validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
				assert.NotNil(tt, err)
				assert.ErrorContains(t, err, "unexpected prefix")
			},
		},
		{
			name:     "comment is not valid yaml",
			document: COMMENT_WITH_INVALID_YAML,
			validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, "yaml: unmarshal errors")
			},
		},
		{
			name:     "comment has no jsonschema properties",
			document: DOESNT_SET_SCHEMA_PROPERTIES,
			validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, *s, jsonschema.Schema{})
			},
		},
		{
			name:     "comment sets jsonschema field: default",
			document: SETS_SCHEMA_DEFAULT,
			validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, s.Default, "baz")
			},
		},
		{
			name:     "comment sets jsonschema field w/ multiline value",
			document: SETS_SCHEMA_WITH_MULTILINE_VALUE,
			validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
				assert.NoError(tt, err)
				assert.Equal(tt, s.Default, "foo\nbar")
			},
		},
		// {
		// 	name:     "comment sets jsonschema description to second yaml doc",
		// 	document: SETS_SCHEMA_WITH_MULTILINE_VALUE,
		// 	validate: func(tt *testing.T, s *jsonschema.Schema, err error) {
		// 		assert.NoError(tt, err)
		// 		assert.Equal(tt, s.Default, "baz")
		// 		assert.Equal(tt, s.Description, "this is a description")
		// 	},
		// },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			yamlNode := &yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.document), yamlNode)
			assert.NoError(tt, err)

			s := &jsonschema.Schema{}
			err = updateSchmeaFromYamlComment(getYamlNode(yamlNode), s)

			tc.validate(tt, s, err)
		})
	}
}

func getYamlNode(node *yaml.Node) *yaml.Node {
	if len(node.Content) == 0 {
		return node
	}

	// Get the first scalar node in the document
	return node.Content[0].Content[0]
}
