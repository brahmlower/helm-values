package comments_test

import (
	"strings"
	"testing"

	"helmvalues/pkg/schema/comments"

	"go.yaml.in/yaml/v4"
)

// FuzzParseHeadComment guards against the class of bug this package has
// twice shipped: comment text that panics or otherwise crashes Parse
// instead of either succeeding or returning a proper error (a
// duplicate-mapping-key nil pointer panic, and a colon-in-description
// heuristic that silently dropped data). Fuzzing only proves crash-safety —
// it has no oracle for "this description should have survived", so it
// complements rather than replaces the hand-written boundary tests above.
func FuzzParseHeadComment(f *testing.F) {
	seeds := []string{
		"comment is just a string",
		"Contact us: support@example.com",
		"See docs for details: https://example.com",
		"Note: this is important",
		"Time format: HH:MM:SS",
		"default: baz",
		"type: string",
		"",
		":",
		"::::",
		"a: b\n---\nc: d",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, raw string) {
		lines := strings.Split(raw, "\n")
		for i, line := range lines {
			lines[i] = "# " + line
		}

		node := &yaml.Node{HeadComment: strings.Join(lines, "\n")}

		// Parse must never panic on arbitrary comment text, regardless of
		// whether it returns a valid schema or an error.
		_, _ = comments.Parse(node, nil)
	})
}
