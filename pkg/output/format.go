package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/Bisman-Singh/helm-blame/pkg/blame"
)

const (
	FormatTable = "table"
	FormatJSON  = "json"
)

// Render writes the blame result in the specified format to the writer.
func Render(w io.Writer, result blame.BlameResult, format string, showShadowed bool) error {
	switch format {
	case FormatJSON:
		return renderJSON(w, result)
	case FormatTable:
		return renderTable(w, result, showShadowed)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func renderTable(w io.Writer, result blame.BlameResult, showShadowed bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tw, "KEY\tVALUE\tSOURCE\n")
	fmt.Fprintf(tw, "---\t-----\t------\n")

	for _, entry := range result.Entries {
		valueStr := formatValue(entry.Value)
		sourceStr := formatSource(entry.Source)

		if showShadowed && len(entry.Shadowed) > 0 {
			sourceStr += " *"
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\n", entry.Key, valueStr, sourceStr)

		if showShadowed {
			for _, s := range entry.Shadowed {
				fmt.Fprintf(tw, "\t\t  ← shadowed: %s\n", formatSource(s))
			}
		}
	}

	return tw.Flush()
}

type jsonEntry struct {
	Key      string       `json:"key"`
	Value    interface{}  `json:"value"`
	Source   jsonSource   `json:"source"`
	Shadowed []jsonSource `json:"shadowed,omitempty"`
}

type jsonSource struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Priority int    `json:"priority"`
}

func renderJSON(w io.Writer, result blame.BlameResult) error {
	entries := make([]jsonEntry, len(result.Entries))

	for i, e := range result.Entries {
		var shadowed []jsonSource
		for _, s := range e.Shadowed {
			shadowed = append(shadowed, toJSONSource(s))
		}

		entries[i] = jsonEntry{
			Key:      e.Key,
			Value:    e.Value,
			Source:   toJSONSource(e.Source),
			Shadowed: shadowed,
		}
	}

	output := struct {
		Chart   string      `json:"chart"`
		Count   int         `json:"count"`
		Entries []jsonEntry `json:"entries"`
	}{
		Chart:   result.ChartName,
		Count:   len(entries),
		Entries: entries,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func toJSONSource(s blame.Source) jsonSource {
	return jsonSource{
		Type:     s.Type.String(),
		Path:     s.Path,
		Priority: s.Priority,
	}
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if len(val) > 40 {
			return val[:37] + "..."
		}
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case nil:
		return "<nil>"
	default:
		s := fmt.Sprintf("%v", val)
		if len(s) > 40 {
			return s[:37] + "..."
		}
		return s
	}
}

func formatSource(s blame.Source) string {
	label := s.Type.String()
	path := s.Path

	// Shorten long paths
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		if len(parts) > 3 {
			path = ".../" + strings.Join(parts[len(parts)-2:], "/")
		}
	}

	return fmt.Sprintf("%s (%s)", path, label)
}
