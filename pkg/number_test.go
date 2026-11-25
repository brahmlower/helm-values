package pkg_test

import (
	"encoding/json"
	"testing"

	"helmvalues/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

const (
	half         = "0.5"
	errNotFinite = "not a finite number"
	errNotNumber = "cannot construct !!str"
)

type numberHolder struct {
	Min *pkg.Number `yaml:"minimum"`
}

func TestNumberUnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in          string
		wantJSON    string
		errContains string
	}{
		{in: "0", wantJSON: "0", errContains: ""},
		{in: "-3", wantJSON: "-3", errContains: ""},
		{in: half, wantJSON: half, errContains: ""},
		{in: ".5", wantJSON: half, errContains: ""},
		{in: "+1", wantJSON: "1", errContains: ""},
		{in: "1e3", wantJSON: "1000", errContains: ""},
		{in: "1_000", wantJSON: "1000", errContains: ""},
		{in: "0x10", wantJSON: "16", errContains: ""},
		{in: ".inf", wantJSON: "", errContains: errNotFinite},
		{in: "-.inf", wantJSON: "", errContains: errNotFinite},
		{in: ".nan", wantJSON: "", errContains: errNotFinite},
		{in: "abc", wantJSON: "", errContains: errNotNumber},
		{in: `"5"`, wantJSON: "", errContains: errNotNumber},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			var holder numberHolder

			err := yaml.Unmarshal([]byte("minimum: "+tc.in), &holder)
			if tc.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, holder.Min)

			out, err := json.Marshal(holder.Min)
			require.NoError(t, err)
			assert.Equal(t, tc.wantJSON, string(out))
		})
	}
}

func TestNumberUnmarshalYAMLOmitted(t *testing.T) {
	t.Parallel()

	var holder numberHolder

	require.NoError(t, yaml.Unmarshal([]byte("other: 1"), &holder))
	assert.Nil(t, holder.Min)
}
