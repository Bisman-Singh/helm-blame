package blame

import (
	"sort"
	"strings"
)

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
	entries = pruneNulledChildren(entries)
	entries = pruneReplacedListItems(entries, layers)

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

// pruneNulledChildren removes child keys when a parent key's winning
// value is nil. In Helm, setting a key to null deletes the entire subtree.
func pruneNulledChildren(entries []BlameEntry) []BlameEntry {
	// Collect all keys with nil winning values
	nullKeys := make(map[string]bool)
	for _, e := range entries {
		if e.Value == nil {
			nullKeys[e.Key] = true
		}
	}

	if len(nullKeys) == 0 {
		return entries
	}

	var result []BlameEntry
	for _, e := range entries {
		pruned := false
		for nullKey := range nullKeys {
			// If this entry is a child of a nulled key, skip it
			if e.Key != nullKey && strings.HasPrefix(e.Key, nullKey+".") {
				pruned = true
				break
			}
		}
		if !pruned {
			result = append(result, e)
		}
	}
	return result
}

// pruneReplacedListItems removes list elements from lower-priority layers
// when a higher-priority layer replaces the entire list. In Helm, lists
// are replaced entirely, not merged element-by-element.
func pruneReplacedListItems(entries []BlameEntry, layers []Layer) []BlameEntry {
	// Find list prefixes that appear in multiple layers
	// A list is identified by keys containing [N]
	type listInfo struct {
		prefix   string
		maxIndex int
		source   Source
	}

	// For each list prefix, find the highest-priority layer that defines it
	listWinners := make(map[string]listInfo)
	for _, layer := range layers {
		flat := FlattenValues(layer.Values)
		for _, kv := range flat {
			// Find the list prefix (everything before [N])
			bracketIdx := strings.Index(kv.Key, "[")
			if bracketIdx < 0 {
				continue
			}
			prefix := kv.Key[:bracketIdx]

			existing, ok := listWinners[prefix]
			if !ok || layer.Source.Priority > existing.source.Priority {
				listWinners[prefix] = listInfo{
					prefix: prefix,
					source: layer.Source,
				}
			}
		}
	}

	if len(listWinners) == 0 {
		return entries
	}

	var result []BlameEntry
	for _, e := range entries {
		bracketIdx := strings.Index(e.Key, "[")
		if bracketIdx < 0 {
			result = append(result, e)
			continue
		}
		prefix := e.Key[:bracketIdx]
		winner, ok := listWinners[prefix]
		if !ok {
			result = append(result, e)
			continue
		}
		// Keep only entries from the winning source
		if e.Source.Priority == winner.source.Priority && e.Source.Path == winner.source.Path {
			result = append(result, e)
		}
	}
	return result
}
