package internal

import (
	"fmt"
	"helmschema/cmd/helm-values/internal/jsonschema"
	"strings"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

type DisplayLine struct {
	LineNum int
	Content string
}

func NewCommentError(node *yaml.Node, err error) *CommentError {
	return &CommentError{
		Node: node,
		Err:  err,
	}
}

type CommentError struct {
	Filepath string
	Node     *yaml.Node
	Err      error
}

func (e *CommentError) Render() string {
	lines := append(
		strings.Split(e.Node.HeadComment, "\n"),
		fmt.Sprintf("%s: ...", e.Node.Value),
	)

	displayLines := make([]DisplayLine, 0)
	for i, line := range lines {
		// +1 we added the node value to the list of display lines
		displayLines = append(displayLines, DisplayLine{
			LineNum: e.Node.Line - len(lines) + i + 1,
			Content: line,
		})
	}

	// update yaml error with adjusted line number
	if yamlErr, ok := e.Err.(*yaml.TypeError); ok {
		for _, unmarshalErr := range yamlErr.Errors {
			// UnmarshalErrors report line number as 1-indexed
			unmarshalErr.Line = displayLines[unmarshalErr.Line-1].LineNum
		}
	}

	for i, line := range displayLines {
		lines[i] = fmt.Sprintf("%d |  %s", line.LineNum, line.Content)
	}

	return fmt.Sprintf(
		"%s\n----| %s\n%s\n",
		e.Err.Error(),
		e.Filepath,
		strings.Join(lines, "\n"),
	)
}

func (e *CommentError) RenderToLog(logger *logrus.Logger) {
	for _, l := range strings.Split(e.Render(), "\n") {
		logger.Warn(l)
	}
}

func (e *CommentError) Error() string {
	return e.Err.Error()
}

func updateSchmeaFromYamlComment(node *yaml.Node, s *jsonschema.Schema) error {
	if node.HeadComment == "" {
		return nil
	}

	commentLines := strings.Split(node.HeadComment, "\n")
	for i, line := range commentLines {
		after, found := strings.CutPrefix(line, "# ")
		if !found {
			err := fmt.Errorf("unexpected prefix: %s", line)
			return NewCommentError(node, err)
		}
		commentLines[i] = after
	}

	cleanedValue := strings.Join(commentLines, "\n")
	for _, commentDoc := range strings.Split(cleanedValue, "---") {
		// Unmarshal into a yaml node to see what kind of document it is
		commentNode := &yaml.Node{}
		_ = yaml.Unmarshal([]byte(commentDoc), commentNode)

		// Skip empty docs
		if len(commentNode.Content) == 0 {
			continue
		}

		// Unmarshal map docs directly into the schema
		if commentNode.Content[0].Kind == yaml.MappingNode {
			err := yaml.Unmarshal([]byte(commentDoc), s)
			if err != nil {
				return NewCommentError(node, err)
			}
		}

		// If the doc is just a string, set it as the schema description
		if commentNode.Content[0].Kind == yaml.ScalarNode {
			s.Description = commentNode.Content[0].Value
		}
		// If the doc is just a string but has a colon in it, which results
		// in it being yaml parsed as a doc with a single key/value whose
		// key likely has some spaces in it
		if commentNode.Content[0].Kind == yaml.MappingNode &&
			len(commentNode.Content[0].Content) == 2 &&
			strings.Count(commentNode.Content[0].Content[0].Value, " ") > 1 {
			s.Description = commentDoc
		}
	}

	return nil
}
