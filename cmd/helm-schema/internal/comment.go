package internal

import (
	"fmt"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"strings"

	"go.yaml.in/yaml/v4"
)

func updateSchmeaFromYamlComment(node *yaml.Node, s *jsonschema.Schema) error {
	if node.HeadComment == "" {
		return nil
	}

	commentLines := strings.Split(node.HeadComment, "\n")
	for i, line := range commentLines {
		after, found := strings.CutPrefix(line, "# ")
		if !found {
			return fmt.Errorf("expected doc comment to start with '# ', got: %s", line)
		}
		commentLines[i] = after
	}

	cleanedValue := strings.Join(commentLines, "\n")

	err := yaml.Unmarshal([]byte(cleanedValue), s)
	if err != nil {
		return fmt.Errorf("failed to parse doc comment as yaml: %w", err)
	}

	return nil
}
