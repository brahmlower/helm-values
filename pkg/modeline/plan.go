package modeline

// Plan bundles the config for a single modeline write, giving access to the
// derived schema URL, modeline, and target file manager.
type Plan struct {
	cfg *Config
}

// NewPlan creates a Plan from cfg.
func NewPlan(cfg *Config) *Plan {
	return &Plan{
		cfg: cfg,
	}
}

// ValuesSchemaURLForChart looks up the values-schema URL for this plan's
// chart.
func (p *Plan) ValuesSchemaURLForChart() (string, error) {
	return ValuesSchemaURLForChart(p.cfg.ChartRef)
}

// Modeline builds the modeline this plan should write, pointing at schema.
func (p *Plan) Modeline(schema string) *Modeline {
	return p.cfg.PartialModeline.ModelineWithValue(schema)
}

// FileManager loads the target file's FileModelineManager.
func (p *Plan) FileManager() (*FileModelineManager, error) {
	return NewFileModelineManager(p.cfg.TargetFile)
}
