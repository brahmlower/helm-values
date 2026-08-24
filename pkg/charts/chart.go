// Package charts provides types and helpers for locating and parsing
// Helm chart directories.
package charts

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Chart represents a Helm chart located on disk, rooted at a given directory.
type Chart struct {
	rootPath string
	Details  *ChartDetails
}

// NewChart loads a Chart from the given chart root directory, parsing its
// Chart.yaml file into ChartDetails.
func NewChart(chartRoot string) (*Chart, error) {
	chart := &Chart{
		rootPath: chartRoot,
		Details:  nil,
	}

	content, err := os.ReadFile(chart.ChartFilePath())
	if err != nil {
		return nil, fmt.Errorf("read chart file: %w", err)
	}

	details := &ChartDetails{
		Name:        "",
		Description: "",
		Version:     "",
		Annotations: nil,
	}

	err = yaml.Unmarshal(content, details)
	if err != nil {
		return nil, fmt.Errorf("parse chart file: %w", err)
	}

	chart.Details = details

	return chart, nil
}

// RootPath returns the chart's root directory path.
func (c *Chart) RootPath() string {
	return c.rootPath
}

// ChartFilePath returns the path to the chart's Chart.yaml file.
func (c *Chart) ChartFilePath() string {
	return c.rootPath + "/Chart.yaml"
}

// ValuesFilePath returns the path to the chart's values.yaml file.
func (c *Chart) ValuesFilePath() string {
	return c.rootPath + "/values.yaml"
}

// SchemaFilePath returns the path to the chart's values.schema.json file.
func (c *Chart) SchemaFilePath() string {
	return c.rootPath + "/values.schema.json"
}

// ReadmeMdFilePath returns the path to the chart's README.md file.
func (c *Chart) ReadmeMdFilePath() string {
	return c.rootPath + "/README.md"
}

// ReadmeMdTemplateFilePath returns the path to the chart's README.md.gotmpl
// template file.
func (c *Chart) ReadmeMdTemplateFilePath() string {
	return c.rootPath + "/README.md.gotmpl"
}

// ReadmeRstFilePath returns the path to the chart's README.rst file.
func (c *Chart) ReadmeRstFilePath() string {
	return c.rootPath + "/README.rst"
}

// ReadmeRstTemplateFilePath returns the path to the chart's README.rst.gotmpl
// template file.
func (c *Chart) ReadmeRstTemplateFilePath() string {
	return c.rootPath + "/README.rst.gotmpl"
}

// ChartDetails holds the metadata parsed from a chart's Chart.yaml file.
type ChartDetails struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Annotations map[string]string `yaml:"annotations"`
}

// ValuesSchema returns the chart's values-schema annotation, or an empty
// string if it is not set.
func (d *ChartDetails) ValuesSchema() string {
	schemaURL, ok := d.Annotations["values-schema"]
	if !ok || schemaURL == "" {
		return ""
	}

	return schemaURL
}
