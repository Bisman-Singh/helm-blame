package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Bisman-Singh/helm-blame/pkg/blame"
)

func sampleResult() blame.BlameResult {
	return blame.BlameResult{
		ChartName: "test-chart",
		Entries: []blame.BlameEntry{
			{
				Key:   "image.repository",
				Value: "nginx",
				Source: blame.Source{
					Type:     blame.SourceParentDefault,
					Path:     "values.yaml",
					Priority: 1,
				},
			},
			{
				Key:   "replicaCount",
				Value: 5,
				Source: blame.Source{
					Type:     blame.SourceSetFlag,
					Path:     "--set replicaCount=5",
					Priority: 200,
				},
				Shadowed: []blame.Source{
					{Type: blame.SourceParentDefault, Path: "values.yaml", Priority: 1},
					{Type: blame.SourceValueFile, Path: "prod.yaml", Priority: 100},
				},
			},
		},
	}
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, sampleResult(), FormatTable, false)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "image.repository") {
		t.Error("expected image.repository in output")
	}
	if !strings.Contains(output, "nginx") {
		t.Error("expected nginx in output")
	}
	if !strings.Contains(output, "replicaCount") {
		t.Error("expected replicaCount in output")
	}
	if !strings.Contains(output, "CLI override") {
		t.Error("expected CLI override in output")
	}
}

func TestRenderTableWithShadowed(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, sampleResult(), FormatTable, true)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "shadowed") {
		t.Error("expected shadowed entries in output")
	}
	if !strings.Contains(output, "prod.yaml") {
		t.Error("expected prod.yaml as shadowed source")
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, sampleResult(), FormatJSON, false)
	if err != nil {
		t.Fatal(err)
	}

	var parsed struct {
		Chart   string `json:"chart"`
		Count   int    `json:"count"`
		Entries []struct {
			Key    string `json:"key"`
			Source struct {
				Type string `json:"type"`
			} `json:"source"`
			Shadowed []struct {
				Path string `json:"path"`
			} `json:"shadowed"`
		} `json:"entries"`
	}

	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Chart != "test-chart" {
		t.Errorf("expected chart test-chart, got %s", parsed.Chart)
	}
	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}
	if parsed.Entries[1].Source.Type != "CLI override" {
		t.Errorf("expected CLI override, got %s", parsed.Entries[1].Source.Type)
	}
	if len(parsed.Entries[1].Shadowed) != 2 {
		t.Errorf("expected 2 shadowed, got %d", len(parsed.Entries[1].Shadowed))
	}
}

func TestRenderInvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, sampleResult(), "xml", false)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestFormatValueTruncation(t *testing.T) {
	long := strings.Repeat("a", 50)
	result := formatValue(long)
	if len(result) > 40 {
		t.Errorf("expected truncated value, got len %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("expected ... suffix")
	}
}
