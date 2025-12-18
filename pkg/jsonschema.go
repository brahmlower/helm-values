package pkg

import (
	"regexp"

	om "github.com/elliotchance/orderedmap/v3"
)

type JsonSchema struct {
	Location string `json:"location,omitempty" yaml:"location,omitempty"`

	// Draft *Draft `json:"draft,omitempty"`
	Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`

	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	Always          *bool  `json:"always,omitempty" yaml:"always,omitempty"`
	Ref             string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	RecursiveAnchor bool   `json:"recursiveAnchor,omitempty" yaml:"recursiveAnchor,omitempty"`
	RecursiveRef    string `json:"recursiveRef,omitempty" yaml:"recursiveRef,omitempty"`
	DynamicAnchor   string `json:"dynamicAnchor,omitempty" yaml:"dynamicAnchor,omitempty"`
	DynamicRef      string `json:"dynamicRef,omitempty" yaml:"dynamicRef,omitempty"`

	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	Constant []any  `json:"constant,omitempty" yaml:"constant,omitempty"`
	Enum     []any  `json:"enum,omitempty" yaml:"enum,omitempty"`

	Not   *JsonSchema   `json:"not,omitempty"`
	AllOf []*JsonSchema `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf []*JsonSchema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf []*JsonSchema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	If    *JsonSchema   `json:"if,omitempty"`
	Then  *JsonSchema   `json:"then,omitempty"`
	Else  *JsonSchema   `json:"else,omitempty"`

	MinProperties         int64                               `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	MaxProperties         int64                               `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	Required              []string                            `json:"required,omitempty" yaml:"required,omitempty"`
	Properties            *om.OrderedMap[string, *JsonSchema] `json:"properties,omitempty" yaml:"properties,omitempty"`
	PropertyNames         *JsonSchema                         `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`
	PatternProperties     map[*regexp.Regexp]*JsonSchema      `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	AdditionalProperties  any                                 `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Dependencies          map[string]any                      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	DependentRequired     map[string][]string                 `json:"dependentRequired,omitempty" yaml:"dependentRequired,omitempty"`
	DependentSchemas      map[string]*JsonSchema              `json:"dependentSchemas,omitempty" yaml:"dependentSchemas,omitempty"`
	UnevaluatedProperties *JsonSchema                         `json:"unevaluatedProperties,omitempty" yaml:"unevaluatedProperties,omitempty"`

	MinItems         int64         `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems         int64         `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	Items            any           `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalItems  any           `json:"additionalItems,omitempty" yaml:"additionalItems,omitempty"`
	PrefixItems      []*JsonSchema `json:"prefixItems,omitempty" yaml:"prefixItems,omitempty"`
	Contains         *JsonSchema   `json:"contains,omitempty" yaml:"contains,omitempty"`
	MinContains      int64         `json:"minContains,omitempty" yaml:"minContains,omitempty"`
	MaxContains      int64         `json:"maxContains,omitempty" yaml:"maxContains,omitempty"`
	UnevaluatedItems *JsonSchema   `json:"unevaluatedItems,omitempty" yaml:"unevaluatedItems,omitempty"`

	MinLength        int64          `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength        int64          `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern          *regexp.Regexp `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	ContentEncoding  string         `json:"contentEncoding,omitempty" yaml:"contentEncoding,omitempty"`
	ContentMediaType string         `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`
	ContentSchema    *JsonSchema    `json:"contentSchema,omitempty" yaml:"contentSchema,omitempty"`

	Minimum          int64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum int64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	Maximum          int64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum int64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	MultipleOf       int64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`

	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Default     any    `json:"default,omitempty" yaml:"default,omitempty"`
	Comment     string `json:"comment,omitempty" yaml:"comment,omitempty"`
	ReadOnly    bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly   bool   `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Examples    []any  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`

	// Extensions map[string]ExtSchema `json:"extensions,omitempty"`
}

type NodeInspector func(keyPath []*JsonSchema, schema *JsonSchema)

func (s *JsonSchema) WalkProperties(fn ...NodeInspector) {
	s.walkProperties(fn)
}

func (s *JsonSchema) walkProperties(fns []NodeInspector, keyPath ...*JsonSchema) {
	for _, fn := range fns {
		fn(keyPath, s)
	}

	if s.Properties == nil {
		return
	}

	for _, k := range s.Properties.AllFromFront() {
		k.walkProperties(fns, append(keyPath, s)...)
	}
}
