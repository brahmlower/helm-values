package modeline

import (
	"fmt"
	"helmvalues/pkg/helm"

	"github.com/sirupsen/logrus"
)

func NewPartialModeline(program, key string) PartialModeline {
	return PartialModeline{
		Program: program,
		Key:     key,
	}
}

type PartialModeline struct {
	Program string
	Key     string
}

func (m PartialModeline) ProgramAndKey() string {
	return fmt.Sprintf("%s: %s=", m.Program, m.Key)
}

func (m PartialModeline) ModelineWithValue(value string) *Modeline {
	return &Modeline{
		PartialModeline: m,
		Value:           value,
	}
}

func (m PartialModeline) Matches(pm *PartialModeline) bool {
	return m.Program == pm.Program && m.Key == pm.Key
}

func ParseModeline(line string) (*Modeline, error) {
	var program, key, value string
	n, err := fmt.Sscanf(line, "%s: %s=%s", &program, &key, &value)
	if err != nil {
		return nil, fmt.Errorf("line does not match modeline format: %w", err)
	}
	if n != 3 {
		return nil, fmt.Errorf("line does not match modeline format: expected 3 parts, got %d", n)
	}

	return NewModeline(program, key, value), nil
}

func NewModeline(program, key, value string) *Modeline {
	return &Modeline{
		PartialModeline: PartialModeline{
			Program: program,
			Key:     key,
		},
		Value: value,
	}
}

type Modeline struct {
	PartialModeline
	Value string
}

func (m Modeline) String() string {
	return fmt.Sprintf("%s%s", m.ProgramAndKey(), m.Value)
}

func ValuesSchemaForChart(chartRef string, version string) (string, error) {
	chartDetails, err := helm.ChartDetailsFromRef(chartRef, version)
	if err != nil {
		return "", err
	}

	schemaURL := chartDetails.ValuesSchema()
	if schemaURL == "" {
		return "", fmt.Errorf("chart does not have annotations.values-schema defined")
	}

	return schemaURL, nil
}

func WriteModeline(logger *logrus.Logger, cfg *Config) error {
	plan := NewPlan(cfg)

	schemaURL, err := plan.ValuesSchemaForChart()
	if err != nil {
		return fmt.Errorf("failed to get values schema for chart: %w", err)
	}

	modeline := plan.Modeline(schemaURL)

	fm, err := plan.FileManager()
	if err != nil {
		return fmt.Errorf("failed to create file manager: %w", err)
	}

	fm.SetModeline(modeline)

	return fm.Write(cfg.CreateParents)
}
