package internal

import (
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func FindCharts(logger *logrus.Logger, rootDir string) ([]string, error) {
	chartDirectories := []string{}
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		logger.Tracef("Checking path: %s", path)

		chartFileInfo, err := os.Stat(fmt.Sprintf("%s/Chart.yaml", path))
		if err != nil {
			logger.
				WithField("reason", "error").
				WithError(err).
				Tracef("Skipping path: %s", path)
			return nil
		}
		if chartFileInfo.IsDir() {
			logger.
				WithField("reason", "Chart.yaml is a directory").
				Tracef("Skipping path: %s", path)
			return nil
		}

		valuesFileInfo, err := os.Stat(fmt.Sprintf("%s/Chart.yaml", path))
		if err != nil {
			logger.
				WithField("reason", "error").
				WithError(err).
				Tracef("Skipping path: %s", path)
			return nil
		}
		if valuesFileInfo.IsDir() {
			logger.
				WithField("reason", "values.yaml is a directory").
				Tracef("Skipping path: %s", path)
			return nil
		}

		logger.Infof("Found chart: %s", path)

		chartDirectories = append(chartDirectories, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return chartDirectories, nil
}

func NewPlan(
	chartRoot string,
	stdout bool,
	strictComments bool,
	dryRun bool,
) *Plan {
	return &Plan{
		chartRoot:      chartRoot,
		stdout:         stdout,
		strictComments: strictComments,
		dryRun:         dryRun,
	}
}

type Plan struct {
	chartRoot      string
	strictComments bool
	stdout         bool
	dryRun         bool
}

func (p *Plan) LogIntent(logger *logrus.Logger) {
	logger.Debugf("%s: plan: DryRun=%t", p.chartRoot, p.dryRun)
	logger.Debugf("%s: plan: StrictComments=%t", p.chartRoot, p.strictComments)
	logger.Debugf("%s: plan: Stdout=%t", p.chartRoot, p.stdout)
	logger.Debugf("%s: plan: ChartRoot=%s", p.chartRoot, p.chartRoot)
	logger.Debugf("%s: plan: ChartFile=%s", p.chartRoot, p.ChartFilePath())
	logger.Debugf("%s: plan: ValuesFile=%s", p.chartRoot, p.ValuesFilePath())
	logger.Debugf("%s: plan: SchemaFile=%s", p.chartRoot, p.SchemaFilePath())
}

func (p *Plan) ChartRoot() string {
	return p.chartRoot
}

func (p *Plan) ChartFilePath() string {
	return fmt.Sprintf("%s/Chart.yaml", p.chartRoot)
}

func (p *Plan) ValuesFilePath() string {
	return fmt.Sprintf("%s/values.yaml", p.chartRoot)
}

func (p *Plan) SchemaFilePath() string {
	return fmt.Sprintf("%s/values.schema.json", p.chartRoot)
}

func (p *Plan) StdOut() bool {
	return p.stdout
}

func (p *Plan) StrictComments() bool {
	return p.strictComments
}

func (p *Plan) DryRun() bool {
	return p.dryRun
}

func (p *Plan) WriteSchema(schema *jsonschema.Schema) error {
	s, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if p.stdout {
		p.writeToStdout(string(s))
	}

	if !p.dryRun {
		return p.WriteToFile(string(s))
	}

	return nil
}

func (p *Plan) WriteToFile(s string) error {
	// p.logger.Debugf("%s: schema: writing schema file", p.chartDir)
	f, err := os.Create(p.SchemaFilePath())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(s)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plan) writeToStdout(s string) {
	fmt.Println(s)
}
