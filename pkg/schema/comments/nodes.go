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

	// If the doc is just a string but has a colon in it, which results
	// in it being yaml parsed as a doc with a single key/value whose
	// key likely has some spaces in it
	if node.Content[0].Kind == yaml.MappingNode &&
		len(node.Content[0].Content) == 2 &&
		strings.Count(node.Content[0].Content[0].Value, " ") > 1 {
		return KeyValueNodes("description", node.Content[0].Value), true
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

func newDocumentNode(content ...*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: content,
	}
}
