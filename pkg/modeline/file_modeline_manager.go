package modeline

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func NewFileModelineManager(filepath string) (*FileModelineManager, error) {
	manager := &FileModelineManager{
		filepath: filepath,
	}

	if _, err := os.Stat(filepath); err == nil {
		manager.exists = true

		data, err := os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}

		manager.content = string(data)
	}

	return manager, nil
}

type FileModelineManager struct {
	filepath string
	exists   bool
	content  string
}

func (m *FileModelineManager) SetModeline(modeline *Modeline) {
	yamlModelinePrefix := fmt.Sprintf("# %s", modeline.ProgramAndKey())
	yamlModeline := fmt.Sprintf("# %s", modeline.String())

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
	}

	m.content = strings.Join(content, "\n")
}

func (m *FileModelineManager) Write(createParents bool) error {
	return os.WriteFile(m.filepath, []byte(m.content), 0644)
}
