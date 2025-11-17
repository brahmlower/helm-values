package internal

import (
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"io/fs"
	"os"

	"github.com/sirupsen/logrus"
)

func BuildPlan(fsys fs.FS, rootDir string, logger *logrus.Logger) ([]*Plan, error) {
	plans := []*Plan{}
	err := fs.WalkDir(fsys, rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		logger.Tracef("Checking path: %s", path)

		chartFileInfo, err := fs.Stat(fsys, fmt.Sprintf("%s/Chart.yaml", path))
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

		valuesFileInfo, err := fs.Stat(fsys, fmt.Sprintf("%s/Chart.yaml", path))
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

		logger.Debugf("Found chart: %s", path)

		plans = append(plans, &Plan{
			FS:       fsys,
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
	FS             fs.FS
	ChartDir       string
	StrictComments bool
	OutputPath     string
	Stdout         bool
	DryRun         bool
}

func (p *Plan) ChartFilePath() string {
	return fmt.Sprintf("%s/Chart.yaml", p.ChartDir)
}

func (p *Plan) ValuesFilePath() string {
	return fmt.Sprintf("%s/values.yaml", p.ChartDir)
}

func (p *Plan) SetSchemaFilename(filename string) {
	p.OutputPath = fmt.Sprintf("%s/%s", p.ChartDir, filename)
}

func (p *Plan) ReadValuesFile() ([]byte, error) {
	p.Logger.Debugf("Reading values file from path: %s", p.ValuesFilePath())
	return fs.ReadFile(p.FS, p.ValuesFilePath())
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
	if p.OutputPath != "" && !p.DryRun {
		return p.WriteToFile(string(s))
	}

	return nil
}

func (p *Plan) WriteToFile(s string) error {
	p.Logger.Debugf("Writing schema to file path: %s", p.OutputPath)
	f, err := os.Create(p.OutputPath)
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
