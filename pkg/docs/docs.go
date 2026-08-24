package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"text/template"

	"helmvalues/pkg"
	"helmvalues/pkg/charts"
	"helmvalues/pkg/docs/templates"
	"helmvalues/pkg/schema"

	"github.com/sirupsen/logrus"
)

// GenerateDocs generates README documentation for every chart discovered under chartDirs,
// rendering the configured (or default) template with the chart's values schema.
func GenerateDocs(logger *logrus.Logger, cfg *Config, chartDirs []string) error {
	chartsFound, err := charts.Search(logger, chartDirs)
	if err != nil {
		return fmt.Errorf("searching for charts: %w", err)
	}

	plans, err := collectPlans(logger, cfg, chartsFound)
	if err != nil {
		return err
	}

	staticPaths, err := templates.StaticTemplates()
	if err != nil {
		return fmt.Errorf("collecting static templates: %w", err)
	}

	// Iterate through plans, generating the docs for each
	for _, plan := range plans {
		if err := generateChartDoc(logger, cfg, staticPaths, plan); err != nil {
			return err
		}
	}

	return nil
}

// collectPlans builds a Plan for each discovered chart, logging its details and
// verifying that a target template can be resolved for it.
func collectPlans(logger *logrus.Logger, cfg *Config, chartsFound []*charts.Chart) ([]*Plan, error) {
	plans := []*Plan{}

	for _, chart := range chartsFound {
		plan := NewPlan(cfg, chart)

		plan.LogCommonDetails(logger)
		plan.LogChartDetails(logger)
		plan.LogSchemaDetails(logger)
		plan.LogDocDetails(logger)

		if _, _, err := plan.DocsTargetTemplate(); err != nil {
			return nil, fmt.Errorf("default template disallowed, but no template found in chart %s", plan.Chart().RootPath())
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

// generateChartDoc renders and writes the documentation for a single chart's plan,
// using staticPaths and cfg.ExtraTemplates as additional templates.
func generateChartDoc(logger *logrus.Logger, cfg *Config, staticPaths []string, plan *Plan) error {
	logger.Infof("docs: %s: starting generation", plan.Chart().Details.Name)

	logger.Debugf("docs: %s: reading values file", plan.Chart().Details.Name)

	jsonschema, err := schema.NewGenerator(logger, plan.SchemaPlan()).Generate()
	if err != nil {
		logger.Error(err.Error())

		return nil
	}

	logger.Tracef("docs: %s: jsonschema properties: %+v", plan.Chart().Details.Name, jsonschema.Properties)

	table := templates.TemplateContext{
		Raw: &templates.RawContext{
			Chart:  plan.Chart(),
			Values: jsonschema,
		},
		ValuesTable: schemaProperties(jsonschema, cfg.Order, []string{}),
	}

	t, err := buildChartTemplate(logger, cfg, staticPaths, plan)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	logger.Debugf("docs: %s: rendering template", plan.Chart().Details.Name)

	err = t.Execute(buf, table)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	logger.Debugf("docs: %s: writing output", plan.Chart().Details.Name)

	if err := plan.WriteReadme(logger, buf.String()); err != nil {
		return err
	}

	logger.Infof("docs: %s: finished", plan.Chart().Details.Name)

	return nil
}

// buildChartTemplate collects the static, extra, and (if applicable) custom template
// paths for plan and builds the resulting root template.
func buildChartTemplate(
	logger *logrus.Logger,
	cfg *Config,
	staticPaths []string,
	plan *Plan,
) (*template.Template, error) {
	for _, p := range staticPaths {
		logger.Debugf("docs: %s: collecting static template: %s", plan.Chart().Details.Name, p)
	}

	for _, extraTemplate := range cfg.ExtraTemplates {
		logger.Debugf("docs: %s: collecting extra template: %s", plan.Chart().Details.Name, extraTemplate)
	}

	extraTemplates := make([]string, 0, len(staticPaths)+len(cfg.ExtraTemplates))
	extraTemplates = append(extraTemplates, staticPaths...)
	extraTemplates = append(extraTemplates, cfg.ExtraTemplates...)

	if !plan.DocsUseDefault() {
		logger.Debugf(
			"docs: %s: collecting template: %s",
			plan.Chart().Details.Name,
			plan.DocsChartReadmeTemplate(),
		)
	} else {
		logger.Debugf(
			"docs: %s: using builtin default template",
			plan.Chart().Details.Name,
		)
	}

	root, err := os.OpenRoot("/")
	if err != nil {
		return nil, fmt.Errorf("opening root filesystem: %w", err)
	}

	layeredFs := pkg.NewLayeredFS(templates.TemplateFS, root.FS())

	markup, err := plan.DocsMarkup()
	if err != nil {
		return nil, err
	}

	opts := []templates.BuilderOpt{
		templates.WithExtraPaths(extraTemplates),
		templates.WithUseDefault(plan.DocsUseDefault()),
		templates.WithMarkup(markup),
	}
	if !plan.DocsUseDefault() {
		opts = append(opts, templates.WithCustomTemplate(plan.DocsChartReadmeTemplate()))
	}

	builder := templates.NewTemplateBuilder(opts...)

	t, err := builder.Build(layeredFs)
	if err != nil {
		return nil, fmt.Errorf("building template: %w", err)
	}

	return t, nil
}

func schemaProperties(jsonschema *pkg.JsonSchema, order ValuesOrder, parents []string) []templates.ValuesRow {
	rows := []templates.ValuesRow{}

	// Key order is preserved by default
	keys := slices.Collect(jsonschema.Properties.Keys())

	// Sort keys alphabetically if requested
	if order == ValuesOrderAlphabetical {
		sort.Strings(keys)
	}

	for _, key := range keys {
		prop, ok := jsonschema.Properties.Get(key)
		if !ok {
			// should be impossible
			continue
		}

		if prop.Ref != "" {
			rows = append(rows, linkValuesRow(parents, key, fmt.Sprintf("[Ref](%s)", prop.Ref)))

			continue
		}

		if prop.Schema != "" {
			rows = append(rows, linkValuesRow(parents, key, fmt.Sprintf("[Schema](%s)", prop.Schema)))

			continue
		}

		if prop.Type == "object" {
			rows = append(rows, schemaProperties(prop, order, append(parents, key))...)

			continue
		}

		defaultStr, err := json.Marshal(prop.Default)
		if err != nil {
			// TODO: Handle this error better
			fmt.Printf("Error marshaling default value for key %s: %v\n", key, err)
		}

		typeValue := formatEnumType(prop.Type, prop.Enum)

		row := templates.ValuesRow{
			Key:         strings.Join(append(parents, key), "."),
			Type:        typeValue,
			Default:     string(defaultStr),
			Description: prop.Description,
		}
		rows = append(rows, row)
	}

	return rows
}

// linkValuesRow builds a ValuesRow for a property that links out to a referenced
// schema instead of describing a value inline (a $ref or $schema property).
func linkValuesRow(parents []string, key, linkType string) templates.ValuesRow {
	return templates.ValuesRow{
		Key:         strings.Join(append(parents, key), "."),
		Type:        linkType,
		Default:     "",
		Description: "",
	}
}

// formatEnumType renders a property's type alongside its enum values, if any, for
// display in the generated values table.
func formatEnumType(propType string, enum []any) string {
	if len(enum) == 0 {
		return propType
	}

	enumItems := make([]string, len(enum))

	for i, enumItem := range enum {
		enumBytes, err := json.Marshal(enumItem)
		if err != nil {
			// TODO: Handle this error better
			continue
		}

		enumItems[i] = string(enumBytes)
	}

	return fmt.Sprintf(
		"%s (enum)\n%s",
		propType,
		strings.Join(enumItems, ", "),
	)
}
