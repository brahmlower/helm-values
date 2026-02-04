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

	displayFile := NewDisplayFile(e.Filepath)

	for i, line := range lines {
		// +1 because we added the node value to the list of display lines
		lineNumber := e.Node.Line - len(lines) + i + 1
		displayFile.AddLine(lineNumber, line)
	}

	// update yaml error with adjusted line number
	if yamlErr, ok := e.Err.(*yaml.LoadErrors); ok {
		for _, unmarshalErr := range yamlErr.Errors {
			// UnmarshalErrors report line number as 1-indexed
			unmarshalErr.Line = displayFile.Lines()[unmarshalErr.Line-1].LineNum
		}
	}

	newLines := []string{e.Err.Error(), ""}
	return strings.Join(append(newLines, displayFile.Render()...), "\n")
}

func (e *CommentError) RenderToLog(logger *logrus.Logger) {
	for _, l := range strings.Split(e.Render(), "\n") {
		logger.Warn(l)
	}
}

func (e *CommentError) Error() string {
	return e.Err.Error()
}

func NewDisplayFile(filepath string) *displayFile {
	return &displayFile{
		filepath:  filepath,
		lines:     []displayLine{},
		lcolWidth: 0,
		rcolWidth: 0,
	}
}

type displayFile struct {
	filepath  string
	lines     []displayLine
	lcolWidth int
	rcolWidth int
}

func (df *displayFile) Lines() []displayLine {
	return df.lines
}

func (df *displayFile) paddedLeftColWidth() int {
	return df.lcolWidth + 2
}

func (df *displayFile) paddedRightColWidth() int {
	return df.rcolWidth + 1
}

func (df *displayFile) updateLeftColWidth(value int) {
	if value > df.lcolWidth {
		df.lcolWidth = value
	}
}

func (df *displayFile) updateRightColWidth(value int) {
	if value > df.rcolWidth {
		df.rcolWidth = value
	}
}

func (df *displayFile) AddLine(lineNum int, content string) {
	df.lines = append(df.lines, displayLine{
		LineNum: lineNum,
		Content: content,
	})
	df.updateLeftColWidth(len(fmt.Sprintf("%d", lineNum)))
	df.updateRightColWidth(len(content))
}

func (df *displayFile) renderLeftCol(lineNum int) string {
	padding := strings.Repeat(" ", df.paddedLeftColWidth()-len(fmt.Sprintf("%d", lineNum)))
	return fmt.Sprintf("%d%s", lineNum, padding)
}

func (df *displayFile) renderRightCol(content string) string {
	return fmt.Sprintf(" %s", content)
}

func (df *displayFile) checkFilepathLength() {
	if len(df.filepath) > len(df.headerLine()) {
		df.updateRightColWidth(len(df.filepath) - (df.paddedLeftColWidth() + 2))
	}
}

func (df *displayFile) headerLine() string {
	return fmt.Sprintf("%s|%s",
		strings.Repeat("-", df.paddedLeftColWidth()),
		strings.Repeat("-", df.paddedRightColWidth()),
	)
}

func (df *displayFile) Render() []string {
	df.checkFilepathLength()

	output := []string{
		df.filepath,
	}

	output = append(output, df.headerLine())

	for _, line := range df.lines {
		lcol := df.renderLeftCol(line.LineNum)
		rcol := df.renderRightCol(line.Content)
		output = append(output, fmt.Sprintf("%s|%s", lcol, rcol))
	}

	return output
}

type displayLine struct {
	LineNum int
	Content string
}

func lineNumWidth(lines []displayLine) int {
	width := 0
	for _, line := range lines {
		if len(fmt.Sprintf("%d", line.LineNum)) > width {
			width = len(fmt.Sprintf("%d", line.LineNum))
		}
	}
	return width
}
