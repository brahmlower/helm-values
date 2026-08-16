package modeline

import (
	"errors"
	"fmt"

	"helmvalues/pkg/helm"

	"github.com/sirupsen/logrus"
)

// modelineFieldCount is the number of fields a modeline line scans into:
// program, key, and value.
const modelineFieldCount = 3

// PartialModeline identifies a modeline by its program and key, without a
// value, so it can be matched against or completed into a full Modeline.
type PartialModeline struct {
	Program string
	Key     string
}

// NewPartialModeline creates a PartialModeline identifying a modeline by its
// program and key, without a value.
func NewPartialModeline(program, key string) PartialModeline {
	return PartialModeline{
		Program: program,
		Key:     key,
	}
}

// ProgramAndKey renders the "program: key=" prefix shared by every modeline
// with this program and key.
func (m PartialModeline) ProgramAndKey() string {
	return fmt.Sprintf("%s: %s=", m.Program, m.Key)
}

// ModelineWithValue completes this PartialModeline into a full Modeline
// carrying the given value.
func (m PartialModeline) ModelineWithValue(value string) *Modeline {
	return &Modeline{
		PartialModeline: m,
		Value:           value,
	}
}

// Matches reports whether pm identifies the same program and key as m.
func (m PartialModeline) Matches(pm *PartialModeline) bool {
	return m.Program == pm.Program && m.Key == pm.Key
}

// ParseModeline parses a "program: key=value" modeline line.
func ParseModeline(line string) (*Modeline, error) {
	var program, key, value string

	n, err := fmt.Sscanf(line, "%s: %s=%s", &program, &key, &value)
	if err != nil {
		return nil, fmt.Errorf("line does not match modeline format: %w", err)
	}

	if n != modelineFieldCount {
		return nil, fmt.Errorf("line does not match modeline format: expected %d parts, got %d", modelineFieldCount, n)
	}

	return NewModeline(program, key, value), nil
}

// Modeline is a "program: key=value" comment line embedded in a values file.
type Modeline struct {
	PartialModeline

	Value string
}

// NewModeline creates a Modeline from its program, key, and value.
func NewModeline(program, key, value string) *Modeline {
	return &Modeline{
		PartialModeline: PartialModeline{
			Program: program,
			Key:     key,
		},
		Value: value,
	}
}

// String renders the modeline as a "program: key=value" line.
func (m Modeline) String() string {
	return fmt.Sprintf("%s%s", m.ProgramAndKey(), m.Value)
}

// ValuesSchemaURLForChart looks up the values-schema URL annotated on the
// chart referenced by chartRef.
func ValuesSchemaURLForChart(chartRef *helm.ChartRef) (string, error) {
	chartDetails, err := helm.ChartDetailsFromRef(chartRef)
	if err != nil {
		return "", fmt.Errorf("getting chart details for %s: %w", chartRef, err)
	}

	schemaURL := chartDetails.ValuesSchema()
	if schemaURL == "" {
		return "", errors.New("chart does not have annotations.values-schema defined")
	}

	return schemaURL, nil
}

// WriteModeline writes the values-schema modeline into the target file
// described by cfg.
func WriteModeline(_ *logrus.Logger, cfg *Config) error {
	plan := NewPlan(cfg)

	schemaURL, err := plan.ValuesSchemaURLForChart()
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
