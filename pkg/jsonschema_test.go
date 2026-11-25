package pkg_test

import (
	"encoding/json"
	"reflect"
	"strings"
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

// draft07Keywords are the JSON Schema draft-07 keywords (issue #41).
var draft07Keywords = []string{
	"$schema", "$id", "$ref", "$comment", "definitions",
	"title", "description", "default", "readOnly", "writeOnly", "examples",
	"multipleOf", "maximum", "exclusiveMaximum", "minimum", "exclusiveMinimum",
	"maxLength", "minLength", "pattern",
	"additionalItems", "items", "maxItems", "minItems", "uniqueItems", "contains",
	"maxProperties", "minProperties", "required", "additionalProperties",
	"properties", "patternProperties", "dependencies", "propertyNames",
	"const", "enum", "type", "format", "contentMediaType", "contentEncoding",
	"if", "then", "else", "allOf", "anyOf", "oneOf", "not",
}

// laterDraftKeywords are 2019-09/2020-12 keywords JsonSchema still declares.
// Draft-07 validators ignore them; removing them is a possible follow-up.
var laterDraftKeywords = []string{
	"deprecated", "dependentRequired", "dependentSchemas",
	"unevaluatedProperties", "unevaluatedItems",
	"prefixItems", "minContains", "maxContains", "contentSchema",
}

func tagName(tag string) string {
	name, _, _ := strings.Cut(tag, ",")

	return name
}

// TestJsonSchemaKeywordsAreStandard guards against fields encoded under names
// that aren't JSON Schema keywords (issue #41).
func TestJsonSchemaKeywordsAreStandard(t *testing.T) {
	t.Parallel()

	allowed := map[string]bool{}
	for _, k := range append(append([]string{}, draft07Keywords...), laterDraftKeywords...) {
		allowed[k] = true
	}

	typ := reflect.TypeFor[pkg.JsonSchema]()
	for i := range typ.NumField() {
		field := typ.Field(i)
		keyword := tagName(field.Tag.Get("json"))
		decodeKeyword := tagName(field.Tag.Get("yaml"))

		assert.True(t, allowed[keyword], "field %s has non-standard json keyword %q", field.Name, keyword)
		assert.Equal(t, keyword, decodeKeyword, "field %s yaml tag must match json tag", field.Name)
	}
}

func TestJsonSchemaConstEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value *any
		want  any
	}{
		{name: "null", value: new(any), want: nil},
		{name: "string", value: lo.ToPtr[any]("foo"), want: "foo"},
		{name: "false", value: lo.ToPtr[any](false), want: false},
		{name: "zero", value: lo.ToPtr[any](0), want: float64(0)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := pkg.NewJsonSchema()
			s.Const = tc.value

			out := marshalToMap(t, s)

			value, ok := out["const"]
			require.True(t, ok, "missing const")
			assert.Equal(t, tc.want, value)
		})
	}

	t.Run("unset", func(t *testing.T) {
		t.Parallel()

		assert.NotContains(t, marshalToMap(t, pkg.NewJsonSchema()), "const")
	})
}
