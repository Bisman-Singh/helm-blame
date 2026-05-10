package blame

// Source describes where a Helm value originated from.
type Source struct {
	// Type identifies the category of this value source.
	Type SourceType
	// Path is the file path or flag string that provided this value.
	// Examples: "charts/nginx/values.yaml", "production.yaml", "--set replicaCount=5"
	Path string
	// Priority determines override order. Higher priority wins.
	// Follows Helm's merge precedence:
	//   0 = subchart default
	//   1 = parent chart default
	//   100+ = -f files (100, 101, 102... left to right)
	//   200+ = --set flags (200, 201, 202... left to right)
	Priority int
}

// SourceType categorizes where a value came from.
type SourceType int

const (
	SourceSubchartDefault SourceType = iota
	SourceParentDefault
	SourceValueFile
	SourceSetFlag
)

// String returns a human-readable label for the source type.
func (s SourceType) String() string {
	switch s {
	case SourceSubchartDefault:
		return "subchart default"
	case SourceParentDefault:
		return "chart default"
	case SourceValueFile:
		return "override file"
	case SourceSetFlag:
		return "CLI override"
	default:
		return "unknown"
	}
}

// BlameEntry represents a single key-value pair with its provenance.
type BlameEntry struct {
	// Key is the dot-separated path (e.g. "image.repository", "ingress.hosts[0].host").
	Key string
	// Value is the final resolved value.
	Value interface{}
	// Source is where this value came from (the winning source).
	Source Source
	// Shadowed contains sources that set this key but were overridden.
	Shadowed []Source
}

// BlameResult contains the full provenance analysis for a chart.
type BlameResult struct {
	// Entries is the list of all leaf values with their provenance.
	Entries []BlameEntry
	// ChartName is the name of the root chart being analyzed.
	ChartName string
}
