// Package comments parses the YAML comments above and below a values.yaml
// key into JSON Schema fields (description, examples, and other schema
// keywords), and renders parse errors with source context.
package comments

import (
	"fmt"
	"strings"

	"helmvalues/pkg"

	"go.yaml.in/yaml/v4"
)

// Parse builds a JsonSchema for node from its head and foot YAML comments,
// merged with extraNodes (additional schema fields derived elsewhere, e.g.
// the inferred type).
func Parse(node *yaml.Node, extraNodes []*yaml.Node) (*pkg.JsonSchema, error) {
	// new yaml map node to append the schema field nodes to
	schemaMapNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: extraNodes,
	}

	if node.HeadComment != "" {
		headNodes, err := headCommentNodes(node)
		if err != nil {
			return nil, err
		}

		schemaMapNode.Content = append(schemaMapNode.Content, headNodes...)
	}

	if node.FootComment != "" {
		footNodes, err := footCommentExampleNodes(node)
		if err != nil {
			return nil, err
		}

		schemaMapNode.Content = append(schemaMapNode.Content, footNodes...)
	}

	// User-authored head/foot comment fields (appended above) should override
	// the auto-derived extraNodes (e.g. inferred "type") when both set the same
	// key, rather than producing a duplicate-key YAML error.
	schemaMapNode.Content = dedupeMappingNodesKeepLast(schemaMapNode.Content)

	// marshal to a string and subsequently unmarshal into the schema
	fullSchema, err := yaml.Marshal(newDocumentNode(schemaMapNode))
	if err != nil {
		return nil, fmt.Errorf("marshaling schema node: %w", err)
	}

	s := &pkg.JsonSchema{}
	if err := yaml.Unmarshal(fullSchema, s); err != nil {
		return nil, fmt.Errorf("unmarshaling schema: %w", err)
	}

	return s, nil
}

// dedupeMappingNodesKeepLast takes a flat key/value node slice (as found in a
// yaml.Node MappingNode's Content) and removes earlier key/value pairs whose
// key also appears later in the slice, keeping the last occurrence. This lets
// later-appended nodes (e.g. user-authored comment fields) override
// earlier-appended ones (e.g. auto-derived fields like "type") instead of
// producing a duplicate-mapping-key error when both are marshaled together.
func dedupeMappingNodesKeepLast(nodes []*yaml.Node) []*yaml.Node {
	lastIndexForKey := make(map[string]int, len(nodes)/yamlKeyValuePairSizeComments)

	for i := 0; i < len(nodes)-1; i += yamlKeyValuePairSizeComments {
		lastIndexForKey[nodes[i].Value] = i
	}

	deduped := make([]*yaml.Node, 0, len(nodes))

	for i := 0; i < len(nodes)-1; i += yamlKeyValuePairSizeComments {
		if lastIndexForKey[nodes[i].Value] != i {
			continue
		}

		deduped = append(deduped, nodes[i], nodes[i+1])
	}

	return deduped
}

// yamlKeyValuePairSizeComments is the number of yaml.Node entries that make up
// a single key+value pair when chunking a mapping node's Content slice.
const yamlKeyValuePairSizeComments = 2

// headCommentNodes parses node's head comment into schema field nodes
// (description and/or arbitrary schema keywords).
func headCommentNodes(node *yaml.Node) ([]*yaml.Node, error) {
	commentDocs, err := parseNodeComment(node.HeadComment)
	if err != nil {
		return nil, NewCommentError(node, err)
	}

	nodes := []*yaml.Node{}

	for _, commentDoc := range commentDocs {
		descNodes, ok := commentAsDescriptionNodes(commentDoc)
		if ok {
			nodes = append(nodes, descNodes...)

			continue
		}

		mapNodes, ok := commentAsMapNodes(commentDoc)
		if ok {
			nodes = append(nodes, mapNodes...)
		}
	}

	return nodes, nil
}

// footCommentExampleNodes parses node's foot comment into an "examples"
// schema field node, one example per "---"-separated comment block.
func footCommentExampleNodes(node *yaml.Node) ([]*yaml.Node, error) {
	commentDocs, err := parseNodeComment(node.FootComment)
	if err != nil {
		return nil, NewCommentError(node, err)
	}

	exampleNodeKey := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "examples",
	}

	exampleNodeValue := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: []*yaml.Node{},
	}
	for _, commentDoc := range commentDocs {
		exampleNodeValue.Content = append(exampleNodeValue.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strings.TrimSpace(commentDoc),
		})
	}

	return []*yaml.Node{exampleNodeKey, exampleNodeValue}, nil
}

func parseNodeComment(rawComment string) ([]string, error) {
	// split the comment by double newline
	parts := strings.Split(rawComment, "\n\n")
	if len(parts) > 1 {
		rawComment = parts[len(parts)-1]
	}

	commentLines := strings.Split(rawComment, "\n")
	for i, line := range commentLines {
		// handle case where comment line is just "#"
		if line == "#" {
			line = "# "
		}

		after, found := strings.CutPrefix(line, "# ")
		if !found {
			return nil, fmt.Errorf("unexpected prefix: %s (%d of %d lines)", line, i, len(commentLines))
		}

		commentLines[i] = after
	}

	commentDocs := strings.Split(
		strings.Join(commentLines, "\n"),
		"---",
	)

	return commentDocs, nil
}
