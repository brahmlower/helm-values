package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"helmvalues/pkg"
	"helmvalues/pkg/charts"
	"helmvalues/pkg/schema"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestChart writes a minimal chart to a temp directory, with the given version and
// (optionally empty) annotations block appended to Chart.yaml, and loads it.
func writeTestChart(t *testing.T, version, annotations string) *charts.Chart {
	t.Helper()

	dir := t.TempDir()

	chartYaml := "name: test-chart\ndescription: a test chart\nversion: " + version + "\n" + annotations
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartYaml), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("foo: bar\n"), 0o600))

	chart, err := charts.NewChart(dir)
	require.NoError(t, err)

	return chart
}

// testChartVersion is the chart version used across TestPlanCheckSchema's test cases.
const testChartVersion = "1.2.3"

func TestPlanCheckSchema(t *testing.T) {
	t.Parallel()

	testSchema := pkg.NewJsonSchema()
	testSchema.Title = "test-chart values"

	content, err := json.MarshalIndent(testSchema, "", "  ")
	require.NoError(t, err)

	tests := []struct {
		name         string
		version      string
		annotations  string
		existingFile []byte
		wantProblems int
	}{
		{
			name:         "schema matches, no annotation",
			version:      testChartVersion,
			annotations:  "",
			existingFile: content,
			wantProblems: 0,
		},
		{
			name:         "schema file missing",
			version:      testChartVersion,
			annotations:  "",
			existingFile: nil,
			wantProblems: 1,
		},
		{
			name:         "schema file stale",
			version:      testChartVersion,
			annotations:  "",
			existingFile: []byte("{}"),
			wantProblems: 1,
		},
		{
			name:    "annotation references current version",
			version: testChartVersion,
			annotations: "annotations:\n" +
				"  values-schema: https://example.com/refs/tags/helm-chart-" + testChartVersion + "/values.schema.json\n",
			existingFile: content,
			wantProblems: 0,
		},
		{
			name:    "annotation references stale version",
			version: testChartVersion,
			annotations: "annotations:\n" +
				"  values-schema: https://example.com/refs/tags/helm-chart-1.0.0/values.schema.json\n",
			existingFile: content,
			wantProblems: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			chart := writeTestChart(t, tc.version, tc.annotations)

			if tc.existingFile != nil {
				require.NoError(t, os.WriteFile(chart.SchemaFilePath(), tc.existingFile, 0o600))
			}

			cfg := &schema.Config{
				StdOut:        false,
				Strict:        false,
				DryRun:        false,
				GitAdd:        false,
				WriteModeline: false,
				Check:         true,
				LogLevel:      logrus.WarnLevel,
			}
			plan := schema.NewPlan(cfg, chart)

			problems := plan.CheckSchema(testSchema)
			assert.Len(t, problems, tc.wantProblems)
		})
	}
}
