package jsonschema

import (
	"math/big"
	"regexp"
)

type Schema struct {
	Location string `json:"location,omitempty"`

	// Draft *Draft `json:"draft,omitempty"`
	Schema string `json:"$schema,omitempty"`

	Format string `json:"format,omitempty"`

	Always          *bool  `json:"always,omitempty"`
	Ref             string `json:"$ref,omitempty"`
	RecursiveAnchor bool   `json:"recursiveAnchor,omitempty"`
	RecursiveRef    string `json:"recursiveRef,omitempty"`
	DynamicAnchor   string `json:"dynamicAnchor,omitempty"`
	DynamicRef      string `json:"dynamicRef,omitempty"`

	Type     string        `json:"type,omitempty"`
	Constant []interface{} `json:"const,omitempty"`
	Enum     []interface{} `json:"enum,omitempty"`

	// Not   *Schema   `json:"not,omitempty"`
	AllOf []*Schema `json:"allOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty"`
	// If    *Schema   `json:"if,omitempty"`
	// Then  *Schema   `json:"then,omitempty"`
	// Else  *Schema   `json:"else,omitempty"`

	MinProperties         int                        `json:"minProperties,omitempty"`
	MaxProperties         int                        `json:"maxProperties,omitempty"`
	Required              []string                   `json:"required,omitempty"`
	Properties            map[string]*Schema         `json:"properties,omitempty"`
	PropertyNames         *Schema                    `json:"propertyNames,omitempty"`
	RegexProperties       bool                       `json:"regexProperties,omitempty"`
	PatternProperties     map[*regexp.Regexp]*Schema `json:"patternProperties,omitempty"`
	AdditionalProperties  interface{}                `json:"additionalProperties,omitempty"`
	Dependencies          map[string]interface{}     `json:"dependencies,omitempty"`
	DependentRequired     map[string][]string        `json:"dependentRequired,omitempty"`
	DependentSchemas      map[string]*Schema         `json:"dependentSchemas,omitempty"`
	UnevaluatedProperties *Schema                    `json:"unevaluatedProperties,omitempty"`

	MinItems         int         `json:"minItems,omitempty"`
	MaxItems         int         `json:"maxItems,omitempty"`
	UniqueItems      bool        `json:"uniqueItems,omitempty"`
	Items            interface{} `json:"items,omitempty"`
	AdditionalItems  interface{} `json:"additionalItems,omitempty"`
	PrefixItems      []*Schema   `json:"prefixItems,omitempty"`
	Items2020        *Schema     `json:"items2020,omitempty"`
	Contains         *Schema     `json:"contains,omitempty"`
	ContainsEval     bool        `json:"containsEval,omitempty"`
	MinContains      int         `json:"minContains,omitempty"`
	MaxContains      int         `json:"maxContains,omitempty"`
	UnevaluatedItems *Schema     `json:"unevaluatedItems,omitempty"`

	MinLength       int            `json:"minLength,omitempty"`
	MaxLength       int            `json:"maxLength,omitempty"`
	Pattern         *regexp.Regexp `json:"pattern,omitempty"`
	ContentEncoding string         `json:"contentEncoding,omitempty"`

	ContentMediaType string `json:"contentMediaType,omitempty"`

	ContentSchema *Schema `json:"contentSchema,omitempty"`

	Minimum          *big.Rat `json:"minimum,omitempty"`
	ExclusiveMinimum *big.Rat `json:"exclusiveMinimum,omitempty"`
	Maximum          *big.Rat `json:"maximum,omitempty"`
	ExclusiveMaximum *big.Rat `json:"exclusiveMaximum,omitempty"`
	MultipleOf       *big.Rat `json:"multipleOf,omitempty"`

	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Default     interface{}   `json:"default,omitempty"`
	Comment     string        `json:"comment,omitempty"`
	ReadOnly    bool          `json:"readOnly,omitempty"`
	WriteOnly   bool          `json:"writeOnly,omitempty"`
	Examples    []interface{} `json:"examples,omitempty"`
	Deprecated  bool          `json:"deprecated,omitempty"`

	// Extensions map[string]ExtSchema `json:"extensions,omitempty"`
}
