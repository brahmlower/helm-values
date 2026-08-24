package schema

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"helmvalues/pkg"
	"helmvalues/pkg/schema/comments"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

// JSONSchemaURI is the JSON Schema draft version URI written to the "$schema" field
// of generated schemas.
const JSONSchemaURI = "http://json-schema.org/draft-07/schema#"

// yamlKeyValuePairSize is the number of yaml.Node entries that make up a single
// key+value pair when chunking a mapping node's Content slice.
const yamlKeyValuePairSize = 2

// mappingNodeBaseExtraNodeCount is the number of "type"/"additionalProperties" comment
// nodes always added to a mapping node's extra nodes, used as a preallocation hint.
const mappingNodeBaseExtraNodeCount = 2

// Generator builds a JSON schema from a chart's values file.
type Generator struct {
	logger *logrus.Logger
	plan   *Plan
}

// NewGenerator constructs a Generator for the given plan.
func NewGenerator(logger *logrus.Logger, plan *Plan) *Generator {
	return &Generator{
		logger: logger,
		plan:   plan,
	}
}

// Generate reads and parses the plan's chart values file and builds a JSON schema
// describing its structure.
func (g *Generator) Generate() (*pkg.JsonSchema, error) {
	f, err := os.ReadFile(g.plan.chart.ValuesFilePath())
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %w", err)
	}

	rootNode := &yaml.Node{}

	err = yaml.Unmarshal(f, rootNode)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal values file: %w", err)
	}

	if rootNode.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("expected document node, got %d", rootNode.Kind)
	}

	s, err := g.buildMappingNode(nil, rootNode.Content[0])
	if err != nil {
		return nil, err
	}

	s.Schema = JSONSchemaURI
	g.logger.Tracef("schmea generator, properties: %+v", s.Properties)

	s.WalkProperties(
		g.warnUndocumentedValue,
		g.warnUntypedValue,
	)

	return s, err
}

func (g *Generator) buildScalarNode(key *yaml.Node, value *yaml.Node) (*pkg.JsonSchema, error) {
	valueType, err := yamlTagToSchema(value.Tag)
	if err != nil {
		return nil, err
	}

	extraNodes := []*yaml.Node{}
	if valueType != "null" {
		extraNodes = append(extraNodes, comments.KeyValueNodes("type", valueType)...)
	}

	extraNodes = append(extraNodes, comments.KeyValueNodes("title", key.Value)...)
	extraNodes = append(extraNodes, comments.KeyNodeValueNodes("default", value)...)

	s, err := comments.Parse(key, extraNodes)
	if err != nil {
		cErr := &comments.CommentError{
			Filepath: "",
			Node:     nil,
			Err:      nil,
		}
		if errors.As(err, &cErr) {
			cErr.Filepath = g.plan.chart.ValuesFilePath()
			cErr.RenderToLog(g.logger)
		}

		err := fmt.Errorf("doc comment error: %w", err)
		if g.plan.StrictComments() {
			return nil, err
		}

		g.logger.Warn(err.Error())
	}

	return s, nil
}

// TODO: Finish handling sequences.
func (g *Generator) buildSequenceNode(key *yaml.Node, _ *yaml.Node) (*pkg.JsonSchema, error) {
	extraNodes := []*yaml.Node{}
	extraNodes = append(extraNodes, comments.KeyValueNodes("type", "array")...)

	// Not all objects will have a yaml key node, only set key values if they exist
	if key == nil {
		s := pkg.NewJsonSchema()
		s.Properties = pkg.NewEncodableOrderedMap[string, *pkg.JsonSchema]()

		return s, nil
	}

	extraNodes = append(extraNodes, comments.KeyValueNodes("title", key.Value)...)

	s, err := comments.Parse(key, extraNodes)
	if err != nil {
		cErr := &comments.CommentError{
			Filepath: "",
			Node:     nil,
			Err:      nil,
		}
		if errors.As(err, &cErr) {
			cErr.Filepath = g.plan.chart.ValuesFilePath()
			cErr.RenderToLog(g.logger)
		}

		err := fmt.Errorf("doc comment error: %w", err)
		if g.plan.StrictComments() {
			return nil, err
		}
	}

	return s, nil
}

// buildMappingNodeTitledSchema parses the doc comments attached to key into a schema,
// with a "title" field derived from the key added to extraNodes. This is only relevant
// when the mapping node being built has a yaml key node (i.e. it isn't the root node).
func (g *Generator) buildMappingNodeTitledSchema(key *yaml.Node, extraNodes []*yaml.Node) (*pkg.JsonSchema, error) {
	extraNodes = append(extraNodes, comments.KeyValueNodes("title", key.Value)...)

	s, err := comments.Parse(key, extraNodes)
	if err != nil {
		cErr := &comments.CommentError{
			Filepath: "",
			Node:     nil,
			Err:      nil,
		}
		if errors.As(err, &cErr) {
			cErr.Filepath = g.plan.chart.ValuesFilePath()
			cErr.RenderToLog(g.logger)
		}

		wrappedErr := fmt.Errorf("doc comment error: %w", err)
		if g.plan.StrictComments() {
			return nil, wrappedErr
		}
	}

	return s, nil
}

// buildChildNodeSchema builds the schema for a single key/value pair found in a
// mapping node's Content, dispatching to the appropriate builder based on the
// value node's kind.
func (g *Generator) buildChildNodeSchema(childKey *yaml.Node, childValue *yaml.Node) (*pkg.JsonSchema, error) {
	switch childValue.Kind {
	case yaml.ScalarNode:
		childValueSchema, err := g.buildScalarNode(childKey, childValue)
		if err != nil {
			g.logger.Debugf("Error building scalar node for key %s: %v", childKey.Value, err)

			return nil, err
		}

		return childValueSchema, nil
	case yaml.SequenceNode:
		childValueSchema, err := g.buildSequenceNode(childKey, childValue)
		if err != nil {
			g.logger.Debugf("Error building sequence node for key %s: %v", childKey.Value, err)

			return nil, err
		}

		return childValueSchema, nil
	case yaml.MappingNode:
		childValueSchema, err := g.buildMappingNode(childKey, childValue)
		if err != nil {
			g.logger.Debugf("Error building mapping node for key %s: %v", childKey.Value, err)

			return nil, err
		}

		return childValueSchema, nil
	default:
		// should be impossible
		return nil, fmt.Errorf("unsupported yaml type: %v", childValue.Kind)
	}
}

func (g *Generator) buildMappingNode(key *yaml.Node, value *yaml.Node) (*pkg.JsonSchema, error) {
	extraNodes := make([]*yaml.Node, 0, mappingNodeBaseExtraNodeCount)
	extraNodes = append(extraNodes, comments.KeyValueNodes("type", "object")...)
	extraNodes = append(extraNodes, comments.KeyValueNodes("additionalProperties", "false")...)

	// Not all objects will have a yaml key node, only set key values if they exist
	s := pkg.NewJsonSchema()

	if key != nil {
		var err error

		s, err = g.buildMappingNodeTitledSchema(key, extraNodes)
		if err != nil {
			return nil, err
		}
	}

	s.Properties = pkg.NewEncodableOrderedMap[string, *pkg.JsonSchema]()

	for _, child := range lo.Chunk(value.Content, yamlKeyValuePairSize) {
		childKey := child[0]
		childValue := child[1]

		childValueSchema, err := g.buildChildNodeSchema(childKey, childValue)
		if err != nil {
			return nil, err
		}

		s.Properties.Set(childKey.Value, childValueSchema)
	}

	// If there are no properties described in the docs, allow additional properties by default
	//
	// TODO: This isn't quite right - we should only do this if additionalProperties isn't
	// explicitly set to false, or if $schema or $ref hasn't been set
	if len(slices.Collect(s.Properties.Keys())) == 0 {
		s.AdditionalProperties = true
	}

	return s, nil
}

func yamlTagToSchema(tag string) (string, error) {
	switch tag {
	case "!!str":
		return "string", nil
	case "!!int":
		return "number", nil
	case "!!float":
		return "number", nil
	case "!!bool":
		return "boolean", nil
	// case "!!array":
	// 	return "array", nil
	case "!!map":
		return "object", nil
	case "!!null":
		return "null", nil
	default:
		return "", fmt.Errorf("unsupported yaml tag: %s", tag)
	}
}

func isDocumented(schemaPath []*pkg.JsonSchema) bool {
	if schemaPath[len(schemaPath)-1].Description != "" {
		return true
	}

	for _, s := range schemaPath {
		if s.Ref != "" {
			return true
		}
	}

	return false
}

func (g *Generator) warnUndocumentedValue(keyPath []*pkg.JsonSchema, schema *pkg.JsonSchema) {
	if !isDocumented(append(keyPath, schema)) {
		if schema.Title == "" {
			return
		}

		keyValues := []string{}

		for _, k := range append(keyPath, schema) {
			if k.Title == "" {
				continue
			}

			keyValues = append(keyValues, k.Title)
		}

		g.logger.Warnf("value is undocumented: %s", strings.Join(keyValues, "."))
	}
}

func (g *Generator) warnUntypedValue(keyPath []*pkg.JsonSchema, schema *pkg.JsonSchema) {
	if schema.Title == "" {
		return
	}

	if schema.Type != "" {
		return
	}

	keyValues := []string{}

	for _, k := range append(keyPath, schema) {
		if k.Title == "" {
			continue
		}

		keyValues = append(keyValues, k.Title)
	}

	g.logger.Warnf("value has no type: %s", strings.Join(keyValues, "."))
}
