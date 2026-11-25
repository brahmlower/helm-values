package pkg_test

import (
	"encoding/json"
	"testing"

	"helmvalues/pkg"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// boundKeys are the JSON Schema numeric keywords that must be written out
// when set to 0 (issue #36).
var boundKeys = []string{
	"minProperties", "maxProperties",
	"minItems", "maxItems",
	"minContains", "maxContains",
	"minLength", "maxLength",
	"minimum", "exclusiveMinimum",
	"maximum", "exclusiveMaximum",
	"multipleOf",
}

func marshalToMap(t *testing.T, s *pkg.JsonSchema) map[string]any {
	t.Helper()

	content, err := json.Marshal(s)
	require.NoError(t, err)

	out := map[string]any{}
	require.NoError(t, json.Unmarshal(content, &out))

	return out
}

func TestJsonSchemaMarshalsZeroBounds(t *testing.T) {
	t.Parallel()

	s := pkg.NewJsonSchema()
	s.MinProperties = lo.ToPtr(int64(0))
	s.MaxProperties = lo.ToPtr(int64(0))
	s.MinItems = lo.ToPtr(int64(0))
	s.MaxItems = lo.ToPtr(int64(0))
	s.MinContains = lo.ToPtr(int64(0))
	s.MaxContains = lo.ToPtr(int64(0))
	s.MinLength = lo.ToPtr(int64(0))
	s.MaxLength = lo.ToPtr(int64(0))
	s.Minimum = lo.ToPtr(pkg.Number(0))
	s.ExclusiveMinimum = lo.ToPtr(pkg.Number(0))
	s.Maximum = lo.ToPtr(pkg.Number(0))
	s.ExclusiveMaximum = lo.ToPtr(pkg.Number(0))
	s.MultipleOf = lo.ToPtr(pkg.Number(0))

	out := marshalToMap(t, s)

	for _, key := range boundKeys {
		value, ok := out[key]
		if assert.True(t, ok, "missing %s", key) {
			assert.InDelta(t, 0.0, value, 0, key)
		}
	}
}

func TestJsonSchemaOmitsUnsetBounds(t *testing.T) {
	t.Parallel()

	out := marshalToMap(t, pkg.NewJsonSchema())

	for _, key := range boundKeys {
		assert.NotContains(t, out, key)
	}
}
