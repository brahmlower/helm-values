package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"helmschema/internal"
	"helmschema/internal/charts"
	"helmschema/pkg/docs/templates"
	"helmschema/pkg/schema"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

func GenerateDocs(logger *logrus.Logger, cfg *Config, chartDirs []string) error {
	chartsFound, err := charts.Search(logger, chartDirs)
	if err != nil {
		return err
	}

	// Itterate through plan to set the logger and config
	plans := []*Plan{}
	for _, chart := range chartsFound {
		plan := NewPlan(cfg, chart)

		plan.LogCommonDetails(logger)
		plan.LogChartDetails(logger)
		plan.LogSchemaDetails(logger)
		plan.LogDocDetails(logger)

		if _, _, err := plan.DocsTargetTemplate(); err != nil {
			return fmt.Errorf("default template disallowed, but no template found in chart %s", plan.Chart().RootPath())
		}
		plans = append(plans, plan)
	}

	staticPaths, err := templates.StaticTemplates()
	if err != nil {
		return err
	}

	// Iterate through plans again, this time generating the docs
	for _, plan := range plans {
		logger.Infof("docs: %s: starting generation", plan.Chart().Details.Name)

		logger.Debugf("docs: %s: reading values file", plan.Chart().Details.Name)
		schema, err := schema.NewGenerator(logger, plan.SchemaPlan()).Generate()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		table := templates.TemplateContext{
			Raw: &templates.RawContext{
				Chart:  plan.Chart(),
				Values: schema,
			},
			ValuesTable: schemaProperties(schema, cfg.Order, []string{}),
		}

		for _, p := range staticPaths {
			logger.Debugf("docs: %s: collecting static template: %s", plan.Chart().Details.Name, p)
		}
		for _, extraTemplate := range cfg.ExtraTemplates {
			logger.Debugf("docs: %s: collecting extra template: %s", plan.Chart().Details.Name, extraTemplate)
		}
		extraTemplates := append(staticPaths, cfg.ExtraTemplates...)

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
			return err
		}

		layeredFs := internal.NewLayeredFS(templates.TemplateFS, root.FS())

		markup, err := plan.DocsMarkup()
		if err != nil {
			return err
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
			return err
		}

		buf := new(bytes.Buffer)
		logger.Debugf("docs: %s: rendering template", plan.Chart().Details.Name)
		err = t.Execute(buf, table)
		if err != nil {
			return err
		}

		logger.Debugf("docs: %s: writing output", plan.Chart().Details.Name)
		if err := plan.WriteReadme(logger, buf.String()); err != nil {
			return err
		}

		logger.Infof("docs: %s: finished", plan.Chart().Details.Name)
	}

	return nil
}

func schemaProperties(schema *schema.Schema, order ValuesOrder, parents []string) []templates.ValuesRow {
	rows := []templates.ValuesRow{}

	// Key order is preserved by default
	keys := slices.Collect(schema.Properties.Keys())

	// Sort keys alphabetically if requested
	if order == ValuesOrderAlphabetical {
		sort.Strings(keys)
	}

	for _, key := range keys {
		prop, ok := schema.Properties.Get(key)
		if !ok {
			// should be impossible
			continue
		}

		if prop.Ref != "" {
			row := templates.ValuesRow{
				Key:  strings.Join(append(parents, key), "."),
				Type: fmt.Sprintf("[Ref](%s)", prop.Ref),
			}
			rows = append(rows, row)
			continue
		}

		if prop.Schema != "" {
			row := templates.ValuesRow{
				Key:  strings.Join(append(parents, key), "."),
				Type: fmt.Sprintf("[Schema](%s)", prop.Schema),
			}
			rows = append(rows, row)
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

		typeValue := prop.Type
		if len(prop.Enum) > 0 {
			enumItems := make([]string, len(prop.Enum))
			for i, enumItem := range prop.Enum {
				enumBytes, err := json.Marshal(enumItem)
				if err != nil {
					// TODO: Handle this error better
					continue
				}
				enumItems[i] = string(enumBytes)
			}

			typeValue = fmt.Sprintf(
				"%s (enum)\n%s",
				typeValue,
				strings.Join(enumItems, ", "),
			)
		}

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
