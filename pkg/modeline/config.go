package modeline

// Config holds configuration for the modeline command
type Config struct {
	ChartVersion    string
	ChartRef        string
	TargetFile      string
	CreateParents   bool
	PartialModeline PartialModeline
}
