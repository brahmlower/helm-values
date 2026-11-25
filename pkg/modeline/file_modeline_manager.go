package modeline

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// filePerm is the permission mode used when writing the modeline back to
// the target file.
const filePerm = 0o600

// FileModelineManager reads and rewrites the modeline comment in a file's
// content, without touching the rest of the file.
type FileModelineManager struct {
	filepath string
	exists   bool
	content  string
}

// NewFileModelineManager loads the file at filepath, if it exists, so its
// modeline can be inspected or replaced.
func NewFileModelineManager(filepath string) (*FileModelineManager, error) {
	manager := &FileModelineManager{
		filepath: filepath,
		exists:   false,
		content:  "",
	}

	if _, err := os.Stat(filepath); err == nil {
		manager.exists = true

		data, err := os.ReadFile(filepath) //nolint:gosec // CLI intentionally reads a user-supplied local file path
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", filepath, err)
		}

		manager.content = string(data)
	}

	return manager, nil
}

// SetModeline replaces the existing modeline matching modeline's program and
// key in place, or inserts modeline as a new first line if none is found. A
// newly inserted modeline is followed by a blank line, unless the next line is
// already blank, so it is not attached to the first key's doc comment.
func (m *FileModelineManager) SetModeline(modeline *Modeline) {
	yamlModelinePrefix := "# " + modeline.ProgramAndKey()
	yamlModeline := "# " + modeline.String()

	found := false

	content := strings.Split(m.content, "\n")
	for i, line := range content {
		if strings.HasPrefix(line, yamlModelinePrefix) {
			content[i] = yamlModeline
			found = true

			break
		}
	}

	if !found {
		content = slices.Insert(content, 0, yamlModeline)

		if strings.TrimSpace(content[1]) != "" {
			content = slices.Insert(content, 1, "")
		}
	}

	m.content = strings.Join(content, "\n")
}

// Write persists the current content back to disk.
func (m *FileModelineManager) Write(_ bool) error {
	if err := os.WriteFile(m.filepath, []byte(m.content), filePerm); err != nil {
		return fmt.Errorf("writing %s: %w", m.filepath, err)
	}

	return nil
}
