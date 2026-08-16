package comments

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

// colPadding is the number of padding characters used when sizing the
// display columns rendered by displayFile.
const colPadding = 2

// errRenderPrefixLines is the number of fixed lines (the error message and
// a blank separator line) prepended before the rendered display file.
const errRenderPrefixLines = 2

// fileRenderPrefixLines is the number of fixed lines (the filepath and the
// header divider) prepended before the per-line rendered output.
const fileRenderPrefixLines = 2

// CommentError wraps an error encountered while parsing the YAML comments
// on node, so it can be rendered with source context.
type CommentError struct {
	Filepath string
	Node     *yaml.Node
	Err      error
}

// NewCommentError wraps err as a CommentError for node.
func NewCommentError(node *yaml.Node, err error) *CommentError {
	return &CommentError{
		Node: node,
		Err:  err,
	}
}

// Render formats the underlying error alongside the source lines around
// node, for display to the user.
func (e *CommentError) Render() string {
	lines := append(
		strings.Split(e.Node.HeadComment, "\n"),
		e.Node.Value+": ...",
	)

	displayFile := newDisplayFile(e.Filepath)

	for i, line := range lines {
		// +1 because we added the node value to the list of display lines
		lineNumber := e.Node.Line - len(lines) + i + 1
		displayFile.AddLine(lineNumber, line)
	}

	// update yaml error with adjusted line number
	yamlErr := &yaml.LoadErrors{}
	if errors.As(e.Err, &yamlErr) {
		for _, unmarshalErr := range yamlErr.Errors {
			// UnmarshalErrors report line number as 1-indexed
			unmarshalErr.Line = displayFile.Lines()[unmarshalErr.Line-1].LineNum
		}
	}

	rendered := displayFile.Render()

	newLines := make([]string, 0, errRenderPrefixLines+len(rendered))
	newLines = append(newLines, e.Err.Error(), "")
	newLines = append(newLines, rendered...)

	return strings.Join(newLines, "\n")
}

// RenderToLog writes the rendered error to logger, one line per log call.
func (e *CommentError) RenderToLog(logger *logrus.Logger) {
	for l := range strings.SplitSeq(e.Render(), "\n") {
		logger.Warn(l)
	}
}

func (e *CommentError) Error() string {
	return e.Err.Error()
}

// displayFile renders a set of source lines as a two-column, line-numbered
// listing for display in error output.
type displayFile struct {
	filepath  string
	lines     []displayLine
	lcolWidth int
	rcolWidth int
}

// newDisplayFile creates an empty displayFile for the given source filepath.
func newDisplayFile(filepath string) *displayFile {
	return &displayFile{
		filepath:  filepath,
		lines:     []displayLine{},
		lcolWidth: 0,
		rcolWidth: 0,
	}
}

func (df *displayFile) Lines() []displayLine {
	return df.lines
}

func (df *displayFile) AddLine(lineNum int, content string) {
	df.lines = append(df.lines, displayLine{
		LineNum: lineNum,
		Content: content,
	})
	df.updateLeftColWidth(len(strconv.Itoa(lineNum)))
	df.updateRightColWidth(len(content))
}

func (df *displayFile) Render() []string {
	df.checkFilepathLength()

	output := make([]string, 0, fileRenderPrefixLines+len(df.lines))
	output = append(output, df.filepath)
	output = append(output, df.headerLine())

	for _, line := range df.lines {
		lcol := df.renderLeftCol(line.LineNum)
		rcol := df.renderRightCol(line.Content)
		output = append(output, fmt.Sprintf("%s|%s", lcol, rcol))
	}

	return output
}

func (df *displayFile) paddedLeftColWidth() int {
	return df.lcolWidth + colPadding
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

func (df *displayFile) renderLeftCol(lineNum int) string {
	padding := strings.Repeat(" ", df.paddedLeftColWidth()-len(strconv.Itoa(lineNum)))

	return fmt.Sprintf("%d%s", lineNum, padding)
}

func (df *displayFile) renderRightCol(content string) string {
	return " " + content
}

func (df *displayFile) checkFilepathLength() {
	if len(df.filepath) > len(df.headerLine()) {
		df.updateRightColWidth(len(df.filepath) - (df.paddedLeftColWidth() + colPadding))
	}
}

func (df *displayFile) headerLine() string {
	return fmt.Sprintf("%s|%s",
		strings.Repeat("-", df.paddedLeftColWidth()),
		strings.Repeat("-", df.paddedRightColWidth()),
	)
}

type displayLine struct {
	LineNum int
	Content string
}
