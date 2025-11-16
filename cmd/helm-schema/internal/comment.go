package internal

import (
	"fmt"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"strings"

	"go.yaml.in/yaml/v4"
)

func newComment(node *yaml.Node) *Comment {
	c := &Comment{}
	if node.HeadComment != "" {
		c.node = node
	}

	return c
}

type Comment struct {
	node         *yaml.Node
	cleanedValue string
}

func (c *Comment) Clean() error {
	// early exit when no comment on node exists
	if c.node == nil {
		return nil
	}

	commentLines := strings.Split(c.node.HeadComment, "\n")
	for i, line := range commentLines {
		after, found := strings.CutPrefix(line, "# ")
		if !found {
			// should be impossible
			return fmt.Errorf("expected doc comment to start with '# ', got: %s", line)
		}
		commentLines[i] = after
	}

	c.cleanedValue = strings.Join(commentLines, "\n")
	return nil
}

func (c *Comment) Parse(s *jsonschema.Schema) error {
	err := yaml.Unmarshal([]byte(c.cleanedValue), s)
	if err != nil {
		return fmt.Errorf("failed to parse doc comment as yaml: %w", err)
	}

	return nil
}
