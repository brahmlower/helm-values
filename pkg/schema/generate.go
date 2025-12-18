package schema

import (
	"fmt"
	"helmschema/pkg"
	"helmschema/pkg/schema/comments"
	"os"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

const JsonSchemaURI = "http://json-schema.org/draft-07/schema#"

type Generator struct {
	logger *logrus.Logger
	plan   *Plan
}

func NewGenerator(logger *logrus.Logger, plan *Plan) *Generator {
	return &Generator{
		logger: logger,
		plan:   plan,
	}
}

func (g *Generator) Generate() (*pkg.JsonSchema, error) {
	f, err := os.ReadFile(g.plan.chart.ValuesFilePath())
	if err != nil {
		return nil, err
	}

	rootNode := &yaml.Node{}
	err = yaml.Unmarshal(f, rootNode)
	if err != nil {
		return nil, err
	}

	if rootNode.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("expected document node, got %d", rootNode.Kind)
	}

	s, err := g.buildMappingNode(nil, rootNode.Content[0])
	if err != nil {
		return nil, err
	}
	s.Schema = JsonSchemaURI
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
	extraNodes = append(extraNodes, comments.KeyValueNodes("default", value.Value)...)

	s, err := comments.Parse(key, extraNodes)
	if err != nil {
		if cErr, ok := err.(*comments.CommentError); ok {
			cErr.Filepath = g.plan.chart.ValuesFilePath()
			cErr.RenderToLog(g.logger)
		}

		err := fmt.Errorf("doc comment error: %w", err)
		if g.plan.StrictComments() {
			return nil, err
		} else {
			g.logger.Warn(err.Error())
		}
	}

	return s, nil
}

// TODO: Finish handling sequences
func (g *Generator) buildSequenceNode(key *yaml.Node, _ *yaml.Node) (*pkg.JsonSchema, error) {
	extraNodes := []*yaml.Node{}
	extraNodes = append(extraNodes, comments.KeyValueNodes("type", "array")...)

	// Not all objects will have a yaml key node, only set key values if they exist
	if key == nil {
		s := &pkg.JsonSchema{}
		s.Properties = pkg.NewEncodableOrderedMap[string, *pkg.JsonSchema]()
		return s, nil
	}

	extraNodes = append(extraNodes, comments.KeyValueNodes("title", key.Value)...)

	s, err := comments.Parse(key, extraNodes)
	if err != nil {
		if cErr, ok := err.(*comments.CommentError); ok {
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

func (g *Generator) buildMappingNode(key *yaml.Node, value *yaml.Node) (*pkg.JsonSchema, error) {
	extraNodes := []*yaml.Node{}
	extraNodes = append(extraNodes, comments.KeyValueNodes("type", "object")...)
	extraNodes = append(extraNodes, comments.KeyValueNodes("additionalProperties", "false")...)

	// Not all objects will have a yaml key node, only set key values if they exist
	s := &pkg.JsonSchema{}
	if key != nil {
		extraNodes = append(extraNodes, comments.KeyValueNodes("title", key.Value)...)

		var err error
		s, err = comments.Parse(key, extraNodes)
		if err != nil {
			if cErr, ok := err.(*comments.CommentError); ok {
				cErr.Filepath = g.plan.chart.ValuesFilePath()
				cErr.RenderToLog(g.logger)
			}

			err := fmt.Errorf("doc comment error: %w", err)
			if g.plan.StrictComments() {
				return nil, err
			}
		}
	}
	s.Properties = pkg.NewEncodableOrderedMap[string, *pkg.JsonSchema]()

	for _, child := range lo.Chunk(value.Content, 2) {
		childKey := child[0]
		childValue := child[1]

		var err error
		var childValueSchema *pkg.JsonSchema
		switch childValue.Kind {
		case yaml.ScalarNode:
			childValueSchema, err = g.buildScalarNode(childKey, childValue)
			if err != nil {
				g.logger.Debugf("Error building scalar node for key %s: %v", childKey.Value, err)
				return nil, err
			}
		case yaml.SequenceNode:
			childValueSchema, err = g.buildSequenceNode(childKey, childValue)
			if err != nil {
				g.logger.Debugf("Error building sequence node for key %s: %v", childKey.Value, err)
				return nil, err
			}
		case yaml.MappingNode:
			childValueSchema, err = g.buildMappingNode(childKey, childValue)
			if err != nil {
				g.logger.Debugf("Error building mapping node for key %s: %v", childKey.Value, err)
				return nil, err
			}
		default:
			// should be impossible
			return nil, fmt.Errorf("unsupported yaml type: %v", childValue.Kind)
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
