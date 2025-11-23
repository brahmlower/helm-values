package charts

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

func NewChart(chartRoot string) (*Chart, error) {
	chart := &Chart{
		rootPath: chartRoot,
	}

	content, err := os.ReadFile(chart.ChartFilePath())
	if err != nil {
		return nil, err
	}

	details := &ChartDetails{}
	err = yaml.Unmarshal(content, details)
	if err != nil {
		return nil, err
	}

	chart.Details = details

	return chart, nil
}

type Chart struct {
	rootPath string
	Details  *ChartDetails
}

func (c *Chart) RootPath() string {
	return c.rootPath
}

func (p *Chart) ChartFilePath() string {
	return fmt.Sprintf("%s/Chart.yaml", p.rootPath)
}

func (p *Chart) ValuesFilePath() string {
	return fmt.Sprintf("%s/values.yaml", p.rootPath)
}

func (p *Chart) SchemaFilePath() string {
	return fmt.Sprintf("%s/values.schema.json", p.rootPath)
}

func (p *Chart) ReadmeMdFilePath() string {
	return fmt.Sprintf("%s/README.md", p.rootPath)
}

func (p *Chart) ReadmeMdTemplateFilePath() string {
	return fmt.Sprintf("%s/README.md.gotmpl", p.rootPath)
}

func (p *Chart) ReadmeRstFilePath() string {
	return fmt.Sprintf("%s/README.rst", p.rootPath)
}

func (p *Chart) ReadmeRstTemplateFilePath() string {
	return fmt.Sprintf("%s/README.rst.gotmpl", p.rootPath)
}

type ChartDetails struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}
