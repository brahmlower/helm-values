package docs_test

import (
	"os"
	"path/filepath"
	"testing"

	"helmvalues/pkg/charts"
	"helmvalues/pkg/docs"
	"helmvalues/pkg/docs/templates"

	"github.com/samber/mo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestPlanCheckReadme(t *testing.T) {
	t.Parallel()

	const rendered = "# test-chart\n\nrendered content\n"

	tests := []struct {
		name         string
		existingFile *string
		wantProblems int
	}{
		{name: "readme matches", existingFile: strPtr(rendered), wantProblems: 0},
		{name: "readme stale", existingFile: strPtr("old content\n"), wantProblems: 1},
		{name: "readme missing", existingFile: nil, wantProblems: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			chartYaml := "name: test-chart\ndescription: a test chart\nversion: 1.0.0\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartYaml), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("foo: bar\n"), 0o600))

			chart, err := charts.NewChart(dir)
			require.NoError(t, err)

			if tc.existingFile != nil {
				require.NoError(t, os.WriteFile(chart.ReadmeMdFilePath(), []byte(*tc.existingFile), 0o600))
			}

			cfg := &docs.Config{
				LogLevel:       logrus.WarnLevel,
				StdOut:         false,
				Strict:         false,
				DryRun:         false,
				GitAdd:         false,
				Check:          true,
				UseDefault:     mo.None[bool](),
				Output:         mo.None[string](),
				Template:       "",
				ExtraTemplates: nil,
				Markup:         mo.None[templates.Markup](),
				Order:          docs.ValuesOrderPreserve,
			}
			plan := docs.NewPlan(cfg, chart)

			problems, err := plan.CheckReadme(rendered)
			require.NoError(t, err)
			assert.Len(t, problems, tc.wantProblems)
		})
	}
}
