package comments

import (
	"strings"

	"go.yaml.in/yaml/v4"
)

func commentAsDescriptionNodes(comment string) ([]*yaml.Node, bool) {
	node := &yaml.Node{}
	_ = yaml.Unmarshal([]byte(comment), node)

	if len(node.Content) == 0 {
		return []*yaml.Node{}, false
	}

	// If the doc is just a string, set it as the schema description
	if node.Content[0].Kind == yaml.ScalarNode {
		return KeyValueNodes("description", strings.TrimSpace(comment)), true
	}

	// If the doc is just a string but has a colon in it, it gets yaml
	// parsed as a doc with a single key/value pair instead of a scalar.
	// Every real schema keyword (type, default, examples, ...) is a
	// single word, so any space in the parsed key means this was never
	// meant as a keyword and the whole comment is actually a description.
	if node.Content[0].Kind == yaml.MappingNode &&
		len(node.Content[0].Content) == 2 &&
		strings.Contains(node.Content[0].Content[0].Value, " ") {
		return KeyValueNodes("description", strings.TrimSpace(comment)), true
	}

	return []*yaml.Node{}, false
}

func commentAsMapNodes(comment string) ([]*yaml.Node, bool) {
	node := &yaml.Node{}
	_ = yaml.Unmarshal([]byte(comment), node)

	if len(node.Content) == 0 {
		return []*yaml.Node{}, false
	}

	if node.Content[0].Kind != yaml.MappingNode {
		return []*yaml.Node{}, false
	}

	return node.Content[0].Content, true
}

// KeyValueNodes builds the pair of YAML scalar nodes representing a single
// "key: value" mapping entry.
func KeyValueNodes(key string, value string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}

	return []*yaml.Node{keyNode, valueNode}
}

// KeyNodeValueNodes builds a key node paired with a clone of valueNode, preserving
// its Tag and Style. Use this instead of KeyValueNodes when the value comes from an
// existing scalar node (e.g. a values.yaml default) rather than a literal Go string —
// KeyValueNodes always produces an untagged plain scalar, so re-marshaling and
// re-parsing it lets YAML's implicit typing turn a quoted "3000" into an int or a
// quoted "true" into a bool, silently losing the original string type.
func KeyNodeValueNodes(key string, valueNode *yaml.Node) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	clonedValue := &yaml.Node{
		Kind:  valueNode.Kind,
		Tag:   valueNode.Tag,
		Style: valueNode.Style,
		Value: valueNode.Value,
	}

	return []*yaml.Node{keyNode, clonedValue}
}

func newDocumentNode(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: content,
	}
}
