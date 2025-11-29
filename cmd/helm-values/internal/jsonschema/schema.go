package jsonschema

import (
	"math/big"
	"regexp"
)

type Schema struct {
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

	Not   *Schema   `json:"not,omitempty"`
	AllOf []*Schema `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	If    *Schema   `json:"if,omitempty"`
	Then  *Schema   `json:"then,omitempty"`
	Else  *Schema   `json:"else,omitempty"`

	MinProperties         int                        `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	MaxProperties         int                        `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	Required              []string                   `json:"required,omitempty" yaml:"required,omitempty"`
	Properties            map[string]*Schema         `json:"properties,omitempty" yaml:"properties,omitempty"`
	PropertyNames         *Schema                    `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`
	PatternProperties     map[*regexp.Regexp]*Schema `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	AdditionalProperties  any                        `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Dependencies          map[string]any             `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	DependentRequired     map[string][]string        `json:"dependentRequired,omitempty" yaml:"dependentRequired,omitempty"`
	DependentSchemas      map[string]*Schema         `json:"dependentSchemas,omitempty" yaml:"dependentSchemas,omitempty"`
	UnevaluatedProperties *Schema                    `json:"unevaluatedProperties,omitempty" yaml:"unevaluatedProperties,omitempty"`

	MinItems         int       `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems         int       `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems      bool      `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	Items            any       `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalItems  any       `json:"additionalItems,omitempty" yaml:"additionalItems,omitempty"`
	PrefixItems      []*Schema `json:"prefixItems,omitempty" yaml:"prefixItems,omitempty"`
	Contains         *Schema   `json:"contains,omitempty" yaml:"contains,omitempty"`
	MinContains      int       `json:"minContains,omitempty" yaml:"minContains,omitempty"`
	MaxContains      int       `json:"maxContains,omitempty" yaml:"maxContains,omitempty"`
	UnevaluatedItems *Schema   `json:"unevaluatedItems,omitempty" yaml:"unevaluatedItems,omitempty"`

	MinLength       int            `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength       int            `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern         *regexp.Regexp `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	ContentEncoding string         `json:"contentEncoding,omitempty" yaml:"contentEncoding,omitempty"`

	ContentMediaType string `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`

	ContentSchema *Schema `json:"contentSchema,omitempty" yaml:"contentSchema,omitempty"`

	Minimum          *big.Rat `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum *big.Rat `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	Maximum          *big.Rat `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum *big.Rat `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	MultipleOf       *big.Rat `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`

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

func (s *Schema) WalkProperties(fn func(keyPath []*Schema, schema *Schema), keyPath ...*Schema) {
	fn(keyPath, s)

	for _, k := range s.Properties {
		k.WalkProperties(fn, append(keyPath, s)...)
	}
}
