package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	usageMarker = "Usage:"
	schemaCmd   = "schema"
	docsCmd     = "docs"
)

func writeBadChart(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	chartFile := "apiVersion: v2\nname: bad\nversion: 0.1.0\ndescription: A test chart\n"

	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartFile), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("# minimum: .inf\nbad: 2\n"), 0o600))

	return dir
}

// execute runs the program with args, returning the error and combined output.
func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	out := new(bytes.Buffer)

	program := Program(logger)
	program.SetOut(out)
	program.SetErr(out)
	program.SetArgs(args)

	err := program.Execute()

	return out.String(), err
}

func TestGenerationErrorsDoNotPrintUsage(t *testing.T) {
	t.Parallel()

	cases := map[string][]string{
		schemaCmd: {schemaCmd, "--strict", "--write-modeline=false"},
		docsCmd:   {docsCmd, "--strict", "--dry-run"},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out, err := execute(t, append(args, writeBadChart(t))...)
			require.Error(t, err)
			assert.Contains(t, out, "doc comment error")
			assert.NotContains(t, out, usageMarker)
		})
	}
}

func TestFlagErrorsPrintUsage(t *testing.T) {
	t.Parallel()

	for _, command := range []string{schemaCmd, docsCmd} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			out, err := execute(t, command, "--bogus")
			require.Error(t, err)
			assert.Contains(t, out, usageMarker)
		})
	}
}
