package templates //nolint:testpackage // mdMultiline is unexported with no public entry point, so this file uses white-box testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMdMultiline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "escapes a literal pipe so it can't break a markdown table row",
			input:    `{{ .Values.clientUrl | default .Values.appUrl }}`,
			expected: `{{ .Values.clientUrl \| default .Values.appUrl }}`,
		},
		{
			name:     "converts newlines to line breaks",
			input:    "first line\nsecond line",
			expected: "first line</br>second line",
		},
		{
			name:     "escapes pipes and converts newlines together",
			input:    "a | b\nc | d",
			expected: `a \| b</br>c \| d`,
		},
		{
			name:     "leaves strings without pipes or newlines unchanged",
			input:    "plain description",
			expected: "plain description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, mdMultiline(tt.input))
		})
	}
}
