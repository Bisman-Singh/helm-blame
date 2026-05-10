package blame

import "sort"

// Layer represents a single source of values with its provenance metadata.
type Layer struct {
	Values map[string]interface{}
	Source Source
}

// Analyze takes an ordered list of value layers (lowest priority first)
// and produces a BlameResult showing the provenance of every leaf value.
//
// Layers must be ordered by priority: subchart defaults first,
// parent defaults next, -f files left to right, --set flags last.
// This matches Helm's merge precedence.
func Analyze(chartName string, layers []Layer) BlameResult {
	// Track all sources per key, ordered by priority
	keyHistory := make(map[string][]sourceEntry)

	for _, layer := range layers {
		flat := FlattenValues(layer.Values)
		for _, kv := range flat {
			keyHistory[kv.Key] = append(keyHistory[kv.Key], sourceEntry{
				value:  kv.Value,
				source: layer.Source,
			})
		}
	}

	entries := buildEntries(keyHistory)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	return BlameResult{
		Entries:   entries,
		ChartName: chartName,
	}
}

// sourceEntry pairs a value with the source that provided it.
type sourceEntry struct {
	value  interface{}
	source Source
}

// buildEntries converts the per-key history into BlameEntries.
// The highest-priority source wins. All others become shadowed.
func buildEntries(history map[string][]sourceEntry) []BlameEntry {
	entries := make([]BlameEntry, 0, len(history))

	for key, sources := range history {
		if len(sources) == 0 {
			continue
		}

		// Find the winner (highest priority)
		winner := sources[0]
		for _, s := range sources[1:] {
			if s.source.Priority > winner.source.Priority {
				winner = s
			}
		}

		// Collect shadowed sources (everything except the winner)
		var shadowed []Source
		for _, s := range sources {
			if s.source.Priority != winner.source.Priority || s.source.Path != winner.source.Path {
				shadowed = append(shadowed, s.source)
			}
		}

		entries = append(entries, BlameEntry{
			Key:      key,
			Value:    winner.value,
			Source:   winner.source,
			Shadowed: shadowed,
		})
	}

	return entries
}
