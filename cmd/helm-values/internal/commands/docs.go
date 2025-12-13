package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-values/internal"
	"helmschema/cmd/helm-values/internal/charts"
	"helmschema/cmd/helm-values/internal/config"
	"helmschema/cmd/helm-values/internal/docs"
	"helmschema/cmd/helm-values/internal/jsonschema"
	"os"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Docs(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewDocsConfig()

	cmd := &cobra.Command{
		Use:   "docs [flags] chart_dir [...chart_dir]",
		Short: "Generate values docs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			return generateDocs(logger, cfg, args)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func generateDocs(logger *logrus.Logger, cfg *config.DocsConfig, chartDirs []string) error {
	chartsFound, err := charts.Search(logger, chartDirs)
	if err != nil {
		return err
	}

	// Itterate through plan to set the logger and config
	plans := []*internal.Plan{}
	for _, chart := range chartsFound {
		plan := internal.NewPlan(cfg, nil, chart)

		plan.LogIntent(logger)

		if _, _, err := plan.DocsTargetTemplate(); err != nil {
			return fmt.Errorf("default template disallowed, but no template found in chart %s", plan.Chart().RootPath())
		}
		plans = append(plans, plan)
	}

	staticPaths, err := docs.StaticTemplates()
	if err != nil {
		return err
	}

	// Iterate through plans again, this time generating the docs
	for _, plan := range plans {
		logger.Infof("docs: %s: starting generation", plan.Chart().Details.Name)

		logger.Debugf("docs: %s: reading values file", plan.Chart().Details.Name)
		schema, err := internal.NewGenerator(logger, plan).Generate()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		table := docs.TemplateContext{
			Raw: &docs.RawContext{
				Chart:  plan.Chart(),
				Values: schema,
			},
			ValuesTable: schemaProperties(schema, []string{}),
		}

		for _, p := range staticPaths {
			logger.Debugf("docs: %s: collecting static template: %s", plan.Chart().Details.Name, p)
		}
		extraTemplates, err := cfg.ExtraTemplates()
		if err != nil {
			return err
		}
		for _, extraTemplate := range extraTemplates {
			logger.Debugf("docs: %s: collecting extra template: %s", plan.Chart().Details.Name, extraTemplate)
		}
		extraTemplates = append(staticPaths, extraTemplates...)

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

		layeredFs := docs.NewLayeredFS(docs.TemplateFS, root.FS())

		markup, err := plan.DocsMarkup()
		if err != nil {
			return err
		}

		opts := []docs.BuilderOpt{
			docs.WithExtraPaths(extraTemplates),
			docs.WithUseDefault(plan.DocsUseDefault()),
			docs.WithMarkup(markup),
		}
		if !plan.DocsUseDefault() {
			opts = append(opts, docs.WithCustomTemplate(plan.DocsChartReadmeTemplate()))
		}

		builder := docs.NewTemplateBuilder(opts...)
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

func schemaProperties(schema *jsonschema.Schema, parents []string) []docs.ValuesRow {
	rows := []docs.ValuesRow{}

	keys := lo.Keys(schema.Properties)
	sort.Strings(keys)
	for _, key := range keys {
		prop := schema.Properties[key]
		if prop.Ref != "" {
			row := docs.ValuesRow{
				Key:  strings.Join(append(parents, key), "."),
				Type: fmt.Sprintf("[Ref](%s)", prop.Ref),
			}
			rows = append(rows, row)
			continue
		}

		if prop.Schema != "" {
			row := docs.ValuesRow{
				Key:  strings.Join(append(parents, key), "."),
				Type: fmt.Sprintf("[Schema](%s)", prop.Schema),
			}
			rows = append(rows, row)
			continue
		}

		if prop.Type == "object" {
			rows = append(rows, schemaProperties(prop, append(parents, key))...)
			continue
		}

		defaultStr, err := json.Marshal(prop.Default)
		if err != nil {
			// TODO: Handle this error better
			fmt.Printf("Error marshaling default value for key %s: %v\n", key, err)
		}

		row := docs.ValuesRow{
			Key:         strings.Join(append(parents, key), "."),
			Type:        prop.Type,
			Default:     string(defaultStr),
			Description: prop.Description,
		}
		rows = append(rows, row)
	}

	return rows
}
