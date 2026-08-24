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

	// marshal to a string and subsequently unmarshal into the schema
	fullSchema, err := yaml.Marshal(newDocumentNode(schemaMapNode))
	if err != nil {
		return nil, fmt.Errorf("marshaling schema node: %w", err)
	}

	s := pkg.NewJsonSchema()
	if err := yaml.Unmarshal(fullSchema, s); err != nil {
		return nil, fmt.Errorf("unmarshaling schema: %w", err)
	}

	return s, nil
}

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
