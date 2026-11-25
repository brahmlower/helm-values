// Package pkg contains the shared core types for helmvalues: the JSON Schema
// model, a layered filesystem abstraction, and an ordered map used for
// preserving property order when encoding schemas.
package pkg

import (
	"regexp"

	"go.yaml.in/yaml/v4"
)

const (
	constKeyword = "const"
	yamlNullTag  = "!!null"
	// yamlPairSize is the number of nodes per key/value pair in a mapping node's Content.
	yamlPairSize = 2
)

// JsonSchema is an in-memory representation of a JSON Schema document,
// covering the keywords used to validate and describe Helm values.
//
//nolint:revive // JsonSchema is public API referenced throughout the module; renaming is out of scope here
type JsonSchema struct {
	// Comment is the "$comment" keyword: a note for schema maintainers that
	// validators ignore. It is declared first so it is the first key in the
	// encoded JSON.
	Comment string `json:"$comment,omitempty" yaml:"$comment,omitempty"`

	// Draft *Draft `json:"draft,omitempty"`
	Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`

	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Const is the "const" keyword: the instance must equal *Const. nil means
	// unset; a non-nil pointer to nil means `const: null`.
	Const *any  `json:"const,omitempty" yaml:"const,omitempty"`
	Enum  []any `json:"enum,omitempty"  yaml:"enum,omitempty"`

	Not   *JsonSchema   `json:"not,omitempty"   yaml:"not,omitempty"`
	AllOf []*JsonSchema `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf []*JsonSchema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf []*JsonSchema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	If    *JsonSchema   `json:"if,omitempty"    yaml:"if,omitempty"`
	Then  *JsonSchema   `json:"then,omitempty"  yaml:"then,omitempty"`
	Else  *JsonSchema   `json:"else,omitempty"  yaml:"else,omitempty"`

	MinProperties         *int64                                    `json:"minProperties,omitempty"         yaml:"minProperties,omitempty"`
	MaxProperties         *int64                                    `json:"maxProperties,omitempty"         yaml:"maxProperties,omitempty"`
	Required              []string                                  `json:"required,omitempty"              yaml:"required,omitempty"`
	Properties            *EncodableOrderedMap[string, *JsonSchema] `json:"properties,omitempty"            yaml:"properties,omitempty"`
	PropertyNames         *JsonSchema                               `json:"propertyNames,omitempty"         yaml:"propertyNames,omitempty"`
	PatternProperties     map[*regexp.Regexp]*JsonSchema            `json:"patternProperties,omitempty"     yaml:"patternProperties,omitempty"`
	AdditionalProperties  any                                       `json:"additionalProperties,omitempty"  yaml:"additionalProperties,omitempty"` //nolint:lll // long aligned struct tag block, matches sibling fields
	Dependencies          map[string]any                            `json:"dependencies,omitempty"          yaml:"dependencies,omitempty"`
	DependentRequired     map[string][]string                       `json:"dependentRequired,omitempty"     yaml:"dependentRequired,omitempty"`
	DependentSchemas      map[string]*JsonSchema                    `json:"dependentSchemas,omitempty"      yaml:"dependentSchemas,omitempty"`
	UnevaluatedProperties *JsonSchema                               `json:"unevaluatedProperties,omitempty" yaml:"unevaluatedProperties,omitempty"` //nolint:lll // long aligned struct tag block, matches sibling fields

	MinItems         *int64        `json:"minItems,omitempty"         yaml:"minItems,omitempty"`
	MaxItems         *int64        `json:"maxItems,omitempty"         yaml:"maxItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty"      yaml:"uniqueItems,omitempty"`
	Items            any           `json:"items,omitempty"            yaml:"items,omitempty"`
	AdditionalItems  any           `json:"additionalItems,omitempty"  yaml:"additionalItems,omitempty"`
	PrefixItems      []*JsonSchema `json:"prefixItems,omitempty"      yaml:"prefixItems,omitempty"`
	Contains         *JsonSchema   `json:"contains,omitempty"         yaml:"contains,omitempty"`
	MinContains      *int64        `json:"minContains,omitempty"      yaml:"minContains,omitempty"`
	MaxContains      *int64        `json:"maxContains,omitempty"      yaml:"maxContains,omitempty"`
	UnevaluatedItems *JsonSchema   `json:"unevaluatedItems,omitempty" yaml:"unevaluatedItems,omitempty"`

	MinLength        *int64         `json:"minLength,omitempty"        yaml:"minLength,omitempty"`
	MaxLength        *int64         `json:"maxLength,omitempty"        yaml:"maxLength,omitempty"`
	Pattern          *regexp.Regexp `json:"pattern,omitempty"          yaml:"pattern,omitempty"`
	ContentEncoding  string         `json:"contentEncoding,omitempty"  yaml:"contentEncoding,omitempty"`
	ContentMediaType string         `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`
	ContentSchema    *JsonSchema    `json:"contentSchema,omitempty"    yaml:"contentSchema,omitempty"`

	Minimum          *Number `json:"minimum,omitempty"          yaml:"minimum,omitempty"`
	ExclusiveMinimum *Number `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	Maximum          *Number `json:"maximum,omitempty"          yaml:"maximum,omitempty"`
	ExclusiveMaximum *Number `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	MultipleOf       *Number `json:"multipleOf,omitempty"       yaml:"multipleOf,omitempty"`

	Title       string `json:"title,omitempty"       yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Default     any    `json:"default,omitempty"     yaml:"default,omitempty"`
	ReadOnly    bool   `json:"readOnly,omitempty"    yaml:"readOnly,omitempty"`
	WriteOnly   bool   `json:"writeOnly,omitempty"   yaml:"writeOnly,omitempty"`
	Examples    []any  `json:"examples,omitempty"    yaml:"examples,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty"  yaml:"deprecated,omitempty"`

	// Extensions map[string]ExtSchema `json:"extensions,omitempty"`
}

// NewJsonSchema returns a JsonSchema with every field set to its zero value,
// ready to have only the fields relevant to the caller overridden.
//
//nolint:revive // mirrors the JsonSchema type name
func NewJsonSchema() *JsonSchema {
	return &JsonSchema{
		Comment:               "",
		Schema:                "",
		Format:                "",
		Ref:                   "",
		Type:                  "",
		Const:                 nil,
		Enum:                  nil,
		Not:                   nil,
		AllOf:                 nil,
		AnyOf:                 nil,
		OneOf:                 nil,
		If:                    nil,
		Then:                  nil,
		Else:                  nil,
		MinProperties:         nil,
		MaxProperties:         nil,
		Required:              nil,
		Properties:            nil,
		PropertyNames:         nil,
		PatternProperties:     nil,
		AdditionalProperties:  nil,
		Dependencies:          nil,
		DependentRequired:     nil,
		DependentSchemas:      nil,
		UnevaluatedProperties: nil,
		MinItems:              nil,
		MaxItems:              nil,
		UniqueItems:           false,
		Items:                 nil,
		AdditionalItems:       nil,
		PrefixItems:           nil,
		Contains:              nil,
		MinContains:           nil,
		MaxContains:           nil,
		UnevaluatedItems:      nil,
		MinLength:             nil,
		MaxLength:             nil,
		Pattern:               nil,
		ContentEncoding:       "",
		ContentMediaType:      "",
		ContentSchema:         nil,
		Minimum:               nil,
		ExclusiveMinimum:      nil,
		Maximum:               nil,
		ExclusiveMaximum:      nil,
		MultipleOf:            nil,
		Title:                 "",
		Description:           "",
		Default:               nil,
		ReadOnly:              false,
		WriteOnly:             false,
		Examples:              nil,
		Deprecated:            false,
	}
}

// UnmarshalYAML decodes a schema node. yaml skips unmarshalers and pointer
// allocation for null values, so `const: null` is detected on the node.
func (s *JsonSchema) UnmarshalYAML(node *yaml.Node) error {
	type plain JsonSchema

	err := node.Decode((*plain)(s))
	if err != nil {
		return err //nolint:wrapcheck // keep yaml's LoadErrors unwrapped so line marks survive
	}

	if node.Kind == yaml.AliasNode {
		node = node.Alias
	}

	if node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(node.Content); i += yamlPairSize {
		if node.Content[i].Value == constKeyword && node.Content[i+1].ShortTag() == yamlNullTag {
			s.Const = new(any)
		}
	}

	return nil
}

// NodeInspector is a callback invoked for each schema node visited by
// WalkProperties, receiving the path of ancestor schemas and the node itself.
type NodeInspector func(keyPath []*JsonSchema, schema *JsonSchema)

// WalkProperties recursively visits s and each nested property schema,
// calling every provided NodeInspector for each node encountered.
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
