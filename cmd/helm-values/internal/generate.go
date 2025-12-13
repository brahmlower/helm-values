package internal

import (
	"fmt"
	"helmschema/cmd/helm-values/internal/comment"
	"helmschema/cmd/helm-values/internal/jsonschema"
	"os"
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

func (g *Generator) Generate() (*jsonschema.Schema, error) {
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

	s.WalkProperties(
		g.warnUndocumentedValue,
		g.warnUntypedValue,
	)

	return s, err
}

func (g *Generator) buildScalarNode(key *yaml.Node, value *yaml.Node) (*jsonschema.Schema, error) {
	valueType, err := yamlTagToSchema(value.Tag)
	if err != nil {
		return nil, err
	}

	extraNodes := []*yaml.Node{}
	if valueType != "null" {
		extraNodes = append(extraNodes, comment.KeyValueNodes("type", valueType)...)
	}
	extraNodes = append(extraNodes, comment.KeyValueNodes("title", key.Value)...)
	extraNodes = append(extraNodes, comment.KeyValueNodes("default", value.Value)...)

	s := &jsonschema.Schema{}

	if err := comment.ToSchema(s, key, extraNodes); err != nil {
		if cErr, ok := err.(*comment.CommentError); ok {
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

// TODO: Finish handling sequences
func (g *Generator) buildSequenceNode(key *yaml.Node, _ *yaml.Node) (*jsonschema.Schema, error) {
	extraNodes := []*yaml.Node{}
	extraNodes = append(extraNodes, comment.KeyValueNodes("type", "array")...)

	s := &jsonschema.Schema{}

	// Not all objects will have a yaml key node, only set key values if they exist
	if key != nil {
		extraNodes = append(extraNodes, comment.KeyValueNodes("title", key.Value)...)

		if err := comment.ToSchema(s, key, extraNodes); err != nil {
			if cErr, ok := err.(*comment.CommentError); ok {
				cErr.Filepath = g.plan.chart.ValuesFilePath()
				cErr.RenderToLog(g.logger)
			}

			err := fmt.Errorf("doc comment error: %w", err)
			if g.plan.StrictComments() {
				return nil, err
			}
		}
	}

	return s, nil
}

func (g *Generator) buildMappingNode(key *yaml.Node, value *yaml.Node) (*jsonschema.Schema, error) {
	extraNodes := []*yaml.Node{}
	extraNodes = append(extraNodes, comment.KeyValueNodes("type", "object")...)
	extraNodes = append(extraNodes, comment.KeyValueNodes("additionalProperties", "false")...)

	s := &jsonschema.Schema{}

	// Not all objects will have a yaml key node, only set key values if they exist
	if key != nil {
		extraNodes = append(extraNodes, comment.KeyValueNodes("title", key.Value)...)

		if err := comment.ToSchema(s, key, extraNodes); err != nil {
			if cErr, ok := err.(*comment.CommentError); ok {
				cErr.Filepath = g.plan.chart.ValuesFilePath()
				cErr.RenderToLog(g.logger)
			}

			err := fmt.Errorf("doc comment error: %w", err)
			if g.plan.StrictComments() {
				return nil, err
			}
		}
	}
	s.Properties = make(map[string]*jsonschema.Schema, 0)

	for _, child := range lo.Chunk(value.Content, 2) {
		childKey := child[0]
		childValue := child[1]

		var err error
		var childValueSchema *jsonschema.Schema
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

		s.Properties[childKey.Value] = childValueSchema
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

func isDocumented(schemaPath []*jsonschema.Schema) bool {
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

func (g *Generator) warnUndocumentedValue(keyPath []*jsonschema.Schema, schema *jsonschema.Schema) {
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

func (g *Generator) warnUntypedValue(keyPath []*jsonschema.Schema, schema *jsonschema.Schema) {
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
