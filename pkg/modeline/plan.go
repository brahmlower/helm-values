package modeline

func NewPlan(cfg *Config) *Plan {
	return &Plan{
		cfg: cfg,
	}
}

type Plan struct {
	cfg *Config
}

func (p *Plan) ValuesSchemaForChart() (string, error) {
	return ValuesSchemaForChart(p.cfg.ChartRef, p.cfg.ChartVersion)
}

func (p *Plan) Modeline(schema string) *Modeline {
	return p.cfg.PartialModeline.ModelineWithValue(schema)
}

func (p *Plan) FileManager() (*FileModelineManager, error) {
	return NewFileModelineManager(p.cfg.TargetFile)
}
