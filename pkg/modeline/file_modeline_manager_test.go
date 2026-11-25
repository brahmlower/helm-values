package modeline_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"helmvalues/pkg/modeline"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testModeline = "# yaml-language-server: $schema=values.schema.json"
	valuesFile   = "values.yaml"
)

// setModeline sets the schema modeline in dir's values file and returns the
// resulting file content.
func setModeline(t *testing.T, dir string) string {
	t.Helper()

	manager, err := modeline.NewFileModelineManager(filepath.Join(dir, valuesFile))
	require.NoError(t, err)

	manager.SetModeline(modeline.NewModeline("yaml-language-server", "$schema", "values.schema.json"))
	require.NoError(t, manager.Write(false))

	content, err := fs.ReadFile(os.DirFS(dir), valuesFile)
	require.NoError(t, err)

	return string(content)
}

func TestFileModelineManagerSetModeline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content *string
		want    string
	}{
		{
			name:    "first line is a comment",
			content: lo.ToPtr("# -- d\nreplicas: 1\n"),
			want:    testModeline + "\n\n# -- d\nreplicas: 1\n",
		},
		{
			name:    "first line is a key",
			content: lo.ToPtr("replicas: 1\n"),
			want:    testModeline + "\n\nreplicas: 1\n",
		},
		{
			name:    "first line already blank",
			content: lo.ToPtr("\n# -- d\nreplicas: 1\n"),
			want:    testModeline + "\n\n# -- d\nreplicas: 1\n",
		},
		{
			name:    "existing modeline is replaced in place",
			content: lo.ToPtr("# yaml-language-server: $schema=old.json\n# -- d\nreplicas: 1\n"),
			want:    testModeline + "\n# -- d\nreplicas: 1\n",
		},
		{
			name:    "missing file",
			content: nil,
			want:    testModeline + "\n",
		},
		{
			name:    "empty file",
			content: lo.ToPtr(""),
			want:    testModeline + "\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if tc.content != nil {
				require.NoError(t, os.WriteFile(filepath.Join(dir, valuesFile), []byte(*tc.content), 0o600))
			}

			assert.Equal(t, tc.want, setModeline(t, dir))
		})
	}
}

func TestFileModelineManagerSetModelineIsIdempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, valuesFile), []byte("# -- d\nreplicas: 1\n"), 0o600))

	first := setModeline(t, dir)
	assert.Equal(t, first, setModeline(t, dir))
}
