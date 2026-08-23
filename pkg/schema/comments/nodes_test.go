package comments_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"helmvalues/pkg/schema/comments"
)

// parseValueNode parses a single `key: value` document and returns the value node,
// preserving whatever Tag/Style the YAML decoder resolved for it (e.g. a quoted
// string keeps its !!str tag, distinguishing it from an unquoted number or bool).
func parseValueNode(t *testing.T, document string) *yaml.Node {
	t.Helper()

	root := &yaml.Node{}
	err := yaml.Unmarshal([]byte(document), root)
	require.NoError(t, err)

	mapping := root.Content[0]

	return mapping.Content[1]
}

// KeyNodeValueNodes exists because buildScalarNode (in the parent schema package)
// needs to carry a values.yaml scalar's *original* type into the generated schema's
// "default" field. Passing the bare string through KeyValueNodes instead loses that
// type: a quoted "3000" or "true" gets re-marshaled as a plain scalar, and on
// re-parse YAML's implicit typing turns it into an int or a bool.
func TestKeyNodeValueNodesPreservesScalarType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		document string
		expected any
	}{
		{"quoted numeric string stays a string", `foo: "3000"`, "3000"},
		{"quoted boolean-looking string stays a string", `foo: "true"`, "true"},
		{"quoted boolean-looking string (false) stays a string", `foo: "false"`, "false"},
		{"unquoted integer stays a number", `foo: 3000`, 3000},
		{"unquoted boolean stays a bool", `foo: true`, true},
		{"unquoted plain string stays a string", `foo: auto`, "auto"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			valueNode := parseValueNode(t, tc.document)
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "foo"}

			extraNodes := comments.KeyNodeValueNodes("default", valueNode)
			s, err := comments.Parse(keyNode, extraNodes)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, s.Default)
		})
	}
}
