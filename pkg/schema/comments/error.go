package comments

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

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
		// +1 because we added the node value to the list of display lines
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

type DisplayLine struct {
	LineNum int
	Content string
}
