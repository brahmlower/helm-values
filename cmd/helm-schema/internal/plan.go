package internal

import (
	"encoding/json"
	"fmt"
	"helmschema/cmd/helm-schema/internal/jsonschema"
	"os"

	"github.com/sirupsen/logrus"
)

type Plan struct {
	Logger         *logrus.Logger
	StrictComments bool
	ValuesPath     string
	OutputPath     string
	Stdout         bool
}

func (p *Plan) ReadValuesFile() ([]byte, error) {
	p.Logger.Debugf("Reading values file from path: %s", p.ValuesPath)
	return os.ReadFile(p.ValuesPath)
}

func (p *Plan) WriteSchema(schema *jsonschema.Schema) error {
	s, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if p.Stdout {
		p.writeToStdout(string(s))
	}

	if p.OutputPath != "" {
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
