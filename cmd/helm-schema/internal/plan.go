package internal

import (
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func BuildPlan(rootDir string, logger *logrus.Logger) ([]*Plan, error) {
	plans := []*Plan{}
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

		plans = append(plans, &Plan{
			ChartDir: path,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return plans, nil
}

type Plan struct {
	Logger         *logrus.Logger
	ChartDir       string
	StrictComments bool
	SchemaFilePath string
	Stdout         bool
	DryRun         bool
}

func (p *Plan) LogIntent() {
	p.Logger.Debugf("%s: plan: DryRun=%t", p.ChartDir, p.DryRun)
	p.Logger.Debugf("%s: plan: StrictComments=%t", p.ChartDir, p.StrictComments)
	p.Logger.Debugf("%s: plan: Stdout=%t", p.ChartDir, p.Stdout)
	p.Logger.Debugf("%s: plan: ValuesFile=%s", p.ChartDir, p.ValuesFilePath())
	p.Logger.Debugf("%s: plan: SchemaFile=%s", p.ChartDir, p.SchemaFilePath)
}

func (p *Plan) ChartFilePath() string {
	return fmt.Sprintf("%s/Chart.yaml", p.ChartDir)
}

func (p *Plan) ValuesFilePath() string {
	return fmt.Sprintf("%s/values.yaml", p.ChartDir)
}

func (p *Plan) SetSchemaFilename(filename string) {
	p.SchemaFilePath = fmt.Sprintf("%s/%s", p.ChartDir, filename)
}

func (p *Plan) ReadValuesFile() ([]byte, error) {
	p.Logger.Debugf("%s: schema: reading values file", p.ChartDir)
	return os.ReadFile(p.ValuesFilePath())
}

func (p *Plan) WriteSchema(schema *jsonschema.Schema) error {
	s, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if p.Stdout {
		p.writeToStdout(string(s))
	}

	// TODO: This will write to the wrong path
	if p.SchemaFilePath != "" && !p.DryRun {
		return p.WriteToFile(string(s))
	}

	return nil
}

func (p *Plan) WriteToFile(s string) error {
	p.Logger.Debugf("%s: schema: writing schema file", p.ChartDir)
	f, err := os.Create(p.SchemaFilePath)
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
