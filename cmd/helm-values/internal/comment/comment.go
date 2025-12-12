package comment

import (
	"fmt"
	"helmschema/cmd/helm-values/internal/jsonschema"
	"strings"

	"go.yaml.in/yaml/v4"
)

func ToSchema(s *jsonschema.Schema, node *yaml.Node, extraNodes []*yaml.Node) error {
	// new yaml map node to append the schema field nodes to
	schemaMapNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: extraNodes,
	}

	if node.HeadComment != "" {
		commentDocs, err := parseNodeComment(node)
		if err != nil {
			return err
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

	// marshal to a string and subsequently unmarshal into the schema
	fullSchema, err := yaml.Marshal(newDocumentNode(schemaMapNode))
	if err != nil {
		return err
	}

	return yaml.Unmarshal(fullSchema, s)
}

func parseNodeComment(node *yaml.Node) ([]string, error) {
	targetComment := node.HeadComment

	// split the comment by double newline
	parts := strings.Split(targetComment, "\n\n")
	if len(parts) > 1 {
		targetComment = parts[len(parts)-1]
	}

	commentLines := strings.Split(targetComment, "\n")
	for i, line := range commentLines {
		after, found := strings.CutPrefix(line, "# ")
		if !found {
			err := fmt.Errorf("unexpected prefix: %s (%d of %d lines)", line, i, len(commentLines))
			return nil, NewCommentError(node, err)
		}
		commentLines[i] = after
	}

	commentDocs := strings.Split(
		strings.Join(commentLines, "\n"),
		"---",
	)

	return commentDocs, nil
}
