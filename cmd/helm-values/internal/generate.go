package internal

import (
	"fmt"
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

	s.WalkProperties(func(keyPath []*jsonschema.Schema, schema *jsonschema.Schema) {
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
			g.logger.Warnf("undocumented value: %s", strings.Join(keyValues, "."))
		}
	})

	return s, err
}

func (g *Generator) buildScalarNode(key *yaml.Node, value *yaml.Node) (*jsonschema.Schema, error) {
	valueType, err := yamlTagToSchema(value.Tag)
	if err != nil {
		return nil, err
	}

	s := &jsonschema.Schema{}
	s.Type = valueType
	s.Title = key.Value

	if err := updateSchmeaFromYamlComment(key, s); err != nil {
		if cErr, ok := err.(*CommentError); ok {
			cErr.Filepath = g.plan.chart.ValuesFilePath()
			cErr.RenderToLog(g.logger)
		}

		err := fmt.Errorf("doc comment error: %w", err)
		if g.plan.StrictComments() {
			return nil, err
		}
	}

	s.Title = key.Value
	s.Default = value.Value
	return s, nil
}

// TODO: Finish handling sequences
func (g *Generator) buildSequenceNode(key *yaml.Node, _ *yaml.Node) (*jsonschema.Schema, error) {
	s := &jsonschema.Schema{}
	s.Type = "array"

	// Not all objects will have a yaml key node, only set key values if they exist
	if key != nil {
		s.Title = key.Value
		if err := updateSchmeaFromYamlComment(key, s); err != nil {
			if cErr, ok := err.(*CommentError); ok {
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

	return s, nil
}

func (g *Generator) buildMappingNode(key *yaml.Node, value *yaml.Node) (*jsonschema.Schema, error) {
	s := &jsonschema.Schema{}
	s.AdditionalProperties = false
	s.Type = "object"

	// Not all objects will have a yaml key node, only set key values if they exist
	if key != nil {
		s.Title = key.Value

		if err := updateSchmeaFromYamlComment(key, s); err != nil {
			if cErr, ok := err.(*CommentError); ok {
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
				return nil, err
			}
		case yaml.SequenceNode:
			childValueSchema, err = g.buildSequenceNode(childKey, childValue)
			if err != nil {
				return nil, err
			}
		case yaml.MappingNode:
			childValueSchema, err = g.buildMappingNode(childKey, childValue)
			if err != nil {
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
	default:
		return "", fmt.Errorf("unsupported yaml tag: %s", tag)
	}
}

func isDocumented(schemaPath []*jsonschema.Schema) bool {
	if schemaPath[len(schemaPath)-1].Description != "" {
		// fmt.Printf("Description found: %s\n", schemaPath[len(schemaPath)-1].Description)
		return true
	}

	for _, s := range schemaPath {
		if s.Ref != "" {
			// fmt.Printf("Ref found: %s\n", schemaPath[len(schemaPath)-1].Ref)
			return true
		}
	}

	return false
}
