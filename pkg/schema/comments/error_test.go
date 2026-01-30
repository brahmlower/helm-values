package comments

import (
	"bytes"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

func TestNewCommentError(t *testing.T) {
	node := &yaml.Node{
		Line:        10,
		Value:       "test-value",
		HeadComment: "# test comment",
	}
	err := errors.New("test error")

	commentErr := NewCommentError(node, err)

	assert.NotNil(t, commentErr)
	assert.Equal(t, node, commentErr.Node)
	assert.Equal(t, err, commentErr.Err)
	assert.Empty(t, commentErr.Filepath)
}

func TestCommentError_Error(t *testing.T) {
	expectedMsg := "some error message"
	err := errors.New(expectedMsg)

	commentErr := &CommentError{
		Node: &yaml.Node{},
		Err:  err,
	}

	assert.Equal(t, expectedMsg, commentErr.Error())
}

func TestCommentError_Render(t *testing.T) {
	tests := []struct {
		name            string
		filepath        string
		node            *yaml.Node
		err             error
		expectedContent []string
	}{
		{
			name:     "simple error with single line comment",
			filepath: "/path/to/file.yaml",
			node: &yaml.Node{
				Line:        10,
				Value:       "someValue",
				HeadComment: "# this is a comment",
			},
			err: errors.New("test error"),
			expectedContent: []string{
				"test error",
				"",
				"/path/to/file.yaml",
				"----|--------------------",
				"9   | # this is a comment",
				"10  | someValue: ...",
			},
		},
		{
			name:     "error with long filepath",
			filepath: "/path/to/a/file/with/a/long/path.yaml",
			node: &yaml.Node{
				Line:        10,
				Value:       "someValue",
				HeadComment: "# this is a comment",
			},
			err: errors.New("test error"),
			expectedContent: []string{
				"test error",
				"",
				"/path/to/a/file/with/a/long/path.yaml",
				"----|--------------------------------",
				"9   | # this is a comment",
				"10  | someValue: ...",
			},
		},
		{
			name:     "error with multi-line comment",
			filepath: "values.yaml",
			node: &yaml.Node{
				Line:        15,
				Value:       "key",
				HeadComment: "# line 1\n# line 2\n# line 3",
			},
			err: errors.New("validation failed"),
			expectedContent: []string{
				"validation failed",
				"",
				"values.yaml",
				"----|---------",
				"12  | # line 1",
				"13  | # line 2",
				"14  | # line 3",
				"15  | key: ...",
			},
		},
		{
			name:     "error with empty comment",
			filepath: "test.yaml",
			node: &yaml.Node{
				Line:        5,
				Value:       "value",
				HeadComment: "",
			},
			err: errors.New("empty comment error"),
			expectedContent: []string{
				"empty comment error",
				"",
				"test.yaml",
				"---|-----",
				"5  | value: ...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commentErr := &CommentError{
				Filepath: tt.filepath,
				Node:     tt.node,
				Err:      tt.err,
			}

			result := commentErr.Render()

			for _, expectedLine := range tt.expectedContent {
				assert.Contains(t, result, expectedLine, "Rendered output should contain expected line")
			}
		})
	}
}

func TestCommentError_Render_WithYamlTypeError(t *testing.T) {
	node := &yaml.Node{
		Line:        20,
		Value:       "field",
		HeadComment: "# comment line 1\n# comment line 2",
	}

	// Create a yaml.TypeError
	yamlErr := &yaml.TypeError{
		Errors: []*yaml.UnmarshalError{
			{
				Line: 1, // This is relative to the comment block
				Err:  errors.New("cannot unmarshal string into int"),
			},
		},
	}

	commentErr := &CommentError{
		Filepath: "config.yaml",
		Node:     node,
		Err:      yamlErr,
	}

	result := commentErr.Render()

	// The yaml error should be present
	assert.Contains(t, result, "cannot unmarshal")
	// The line numbers should be adjusted
	assert.Contains(t, result, "18  | # comment line 1")
	assert.Contains(t, result, "19  | # comment line 2")
	assert.Contains(t, result, "20  | field: ...")
	assert.Contains(t, result, "config.yaml")
}

func TestCommentError_RenderToLog_MultipleLines(t *testing.T) {
	node := &yaml.Node{
		Line:        10,
		Value:       "value",
		HeadComment: "# comment 1\n# comment 2\n# comment 3",
	}
	err := errors.New("multi-line test")

	commentErr := &CommentError{
		Filepath: "multi.yaml",
		Node:     node,
		Err:      err,
	}

	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.WarnLevel)

	commentErr.RenderToLog(logger)

	logOutput := buf.String()

	// Verify all comments appear
	assert.Contains(t, logOutput, "comment 1")
	assert.Contains(t, logOutput, "comment 2")
	assert.Contains(t, logOutput, "comment 3")
	assert.Contains(t, logOutput, "multi-line test")
}
