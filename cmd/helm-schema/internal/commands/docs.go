package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-schema/internal"
	"helmschema/cmd/helm-schema/internal/charts"
	"helmschema/cmd/helm-schema/internal/config"
	"helmschema/cmd/helm-schema/internal/docs"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Docs(logger *logrus.Logger) *cobra.Command {
	cfg := config.NewDocsConfig()

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate values docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UpdateLogger(logger); err != nil {
				return err
			}

			return generateDocs(logger, cfg)
		},
	}

	cfg.BindFlags(cmd)

	return cmd
}

func generateDocs(logger *logrus.Logger, cfg *config.DocsConfig) error {
	chartDir, err := cfg.ChartDir()
	if err != nil {
		return err
	}

	chartsFound, err := charts.Search(logger, chartDir)
	if err != nil {
		return err
	}
	logger.Infof("Found %d charts", len(chartsFound))

	// Itterate through plan to set the logger and config
	plans := []*internal.Plan{}
	for _, chart := range chartsFound {
		plan := internal.NewPlan(
			chart,
			cfg.StdOut(),
			cfg.Strict(),
			cfg.DryRun(),
		)
		plan.LogIntent(logger)
		plans = append(plans, plan)
	}

	// Iterate through plans again, this time generating the docs
	for _, plan := range plans {
		logger.Debugf("%s: docs: starting generation", plan.Chart().Details.Name)

		logger.Debugf("%s: docs: reading values file", plan.Chart().Details.Name)
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

		extraTemplates, err := cfg.ExtraTemplates()
		if err != nil {
			return err
		}
		extraTemplates = append(extraTemplates, "templates/md.valuesTable.gotmpl")
		extraTemplates = append(extraTemplates, "templates/rst.valuesTable.gotmpl")

		logger.Debugf(
			"%s: docs: loading template: %s",
			plan.Chart().Details.Name,
			plan.ReadmeTemplateFilePath(),
		)
		for _, extraTemplate := range extraTemplates {
			logger.Debugf("%s: docs: loading extra template: %s", plan.Chart().Details.Name, extraTemplate)
		}

		root, err := os.OpenRoot("/")
		if err != nil {
			return err
		}
		// rootFS := root.FS()
		// f, err := rootFS.Open("Users/brahm.lower/development/helm-kiwix/charts/kiwix/README.md.gotmpl")
		// fmt.Printf("f: %v\n", f)
		// fmt.Printf("err: %s\n", err.Error())

		layeredFs := docs.NewLayeredFS(
			docs.TemplateFS,
			root.FS(),
		)

		t, err := docs.NewTemplator(
			logger,
			layeredFs,
			plan.Chart().Details.Name,
			plan.ReadmeTemplateFilePath(),
			extraTemplates,
		)
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		logger.Debugf("%s: docs: rendering readme file", plan.Chart().Details.Name)
		err = t.Render(buf, table)
		if err != nil {
			return err
		}

		if err := plan.WriteReadme(logger, buf.String()); err != nil {
			return err
		}
	}

	return nil
}

func schemaProperties(schema *jsonschema.Schema, parents []string) []docs.ValuesRow {
	rows := []docs.ValuesRow{}

	for key, prop := range schema.Properties {
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
