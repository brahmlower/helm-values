package comments

import (
	"fmt"
	"helmvalues/pkg"
	"strings"

	"go.yaml.in/yaml/v4"
)

func Parse(node *yaml.Node, extraNodes []*yaml.Node) (*pkg.JsonSchema, error) {
	// new yaml map node to append the schema field nodes to
	schemaMapNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: extraNodes,
	}

	if node.HeadComment != "" {
		commentDocs, err := parseNodeComment(node.HeadComment)
		if err != nil {
			return nil, NewCommentError(node, err)
		}

		for _, commentDoc := range commentDocs {
			nodes, ok := commentAsDescriptionNodes(commentDoc)
			if ok {
				schemaMapNode.Content = append(schemaMapNode.Content, nodes...)
				continue
			}

			nodes, ok = commentAsMapNodes(commentDoc)
			if ok {
				schemaMapNode.Content = append(schemaMapNode.Content, nodes...)
			}
		}
	}

	if node.FootComment != "" {
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

		schemaMapNode.Content = append(
			schemaMapNode.Content,
			exampleNodeKey,
			exampleNodeValue,
		)
	}

	// marshal to a string and subsequently unmarshal into the schema
	fullSchema, err := yaml.Marshal(newDocumentNode(schemaMapNode))
	if err != nil {
		return nil, err
	}

	s := &pkg.JsonSchema{}
	err = yaml.Unmarshal(fullSchema, s)

	return s, err
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
