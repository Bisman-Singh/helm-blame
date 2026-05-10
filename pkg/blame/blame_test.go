package blame

import (
	"testing"
)

func TestAnalyzeSingleLayer(t *testing.T) {
	layers := []Layer{
		{
			Values: map[string]interface{}{
				"replicaCount": 1,
				"image": map[string]interface{}{
					"repository": "nginx",
				},
			},
			Source: Source{Type: SourceParentDefault, Path: "values.yaml", Priority: 1},
		},
	}

	result := Analyze("test-chart", layers)

	if result.ChartName != "test-chart" {
		t.Errorf("expected chart name test-chart, got %s", result.ChartName)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}

	for _, e := range result.Entries {
		if e.Source.Path != "values.yaml" {
			t.Errorf("key %s: expected source values.yaml, got %s", e.Key, e.Source.Path)
		}
		if len(e.Shadowed) != 0 {
			t.Errorf("key %s: expected no shadowed sources", e.Key)
		}
	}
}

func TestAnalyzeOverride(t *testing.T) {
	layers := []Layer{
		{
			Values: map[string]interface{}{
				"replicaCount": 1,
			},
			Source: Source{Type: SourceParentDefault, Path: "values.yaml", Priority: 1},
		},
		{
			Values: map[string]interface{}{
				"replicaCount": 3,
			},
			Source: Source{Type: SourceValueFile, Path: "production.yaml", Priority: 100},
		},
	}

	result := Analyze("my-app", layers)

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	if entry.Value != 3 {
		t.Errorf("expected value 3, got %v", entry.Value)
	}
	if entry.Source.Path != "production.yaml" {
		t.Errorf("expected source production.yaml, got %s", entry.Source.Path)
	}
	if len(entry.Shadowed) != 1 {
		t.Fatalf("expected 1 shadowed source, got %d", len(entry.Shadowed))
	}
	if entry.Shadowed[0].Path != "values.yaml" {
		t.Errorf("expected shadowed source values.yaml, got %s", entry.Shadowed[0].Path)
	}
}

func TestAnalyzeSetFlagWins(t *testing.T) {
	layers := []Layer{
		{
			Values:  map[string]interface{}{"replicaCount": 1},
			Source: Source{Type: SourceParentDefault, Path: "values.yaml", Priority: 1},
		},
		{
			Values:  map[string]interface{}{"replicaCount": 3},
			Source: Source{Type: SourceValueFile, Path: "prod.yaml", Priority: 100},
		},
		{
			Values:  map[string]interface{}{"replicaCount": 5},
			Source: Source{Type: SourceSetFlag, Path: "--set replicaCount=5", Priority: 200},
		},
	}

	result := Analyze("my-app", layers)
	entry := result.Entries[0]

	if entry.Value != 5 {
		t.Errorf("expected value 5, got %v", entry.Value)
	}
	if entry.Source.Type != SourceSetFlag {
		t.Errorf("expected source type CLI override, got %s", entry.Source.Type)
	}
	if len(entry.Shadowed) != 2 {
		t.Errorf("expected 2 shadowed sources, got %d", len(entry.Shadowed))
	}
}

func TestAnalyzeSubchartDefault(t *testing.T) {
	layers := []Layer{
		{
			Values: map[string]interface{}{
				"image": map[string]interface{}{"tag": "latest"},
			},
			Source: Source{Type: SourceSubchartDefault, Path: "charts/nginx/values.yaml", Priority: 0},
		},
		{
			Values: map[string]interface{}{
				"image": map[string]interface{}{"tag": "1.25"},
			},
			Source: Source{Type: SourceParentDefault, Path: "values.yaml", Priority: 1},
		},
	}

	result := Analyze("umbrella", layers)
	entry := result.Entries[0]

	if entry.Key != "image.tag" {
		t.Errorf("expected key image.tag, got %s", entry.Key)
	}
	if entry.Value != "1.25" {
		t.Errorf("expected value 1.25, got %v", entry.Value)
	}
	if entry.Source.Type != SourceParentDefault {
		t.Errorf("expected parent default, got %s", entry.Source.Type)
	}
}

func TestAnalyzeNoShadowWhenOnlyOneSource(t *testing.T) {
	layers := []Layer{
		{
			Values: map[string]interface{}{"a": 1, "b": 2},
			Source: Source{Type: SourceParentDefault, Path: "values.yaml", Priority: 1},
		},
		{
			Values: map[string]interface{}{"c": 3},
			Source: Source{Type: SourceValueFile, Path: "extra.yaml", Priority: 100},
		},
	}

	result := Analyze("test", layers)

	for _, e := range result.Entries {
		if len(e.Shadowed) != 0 {
			t.Errorf("key %s should have no shadowed sources", e.Key)
		}
	}
}

func TestAnalyzeSorted(t *testing.T) {
	layers := []Layer{
		{
			Values: map[string]interface{}{"zebra": 1, "alpha": 2, "middle": 3},
			Source: Source{Type: SourceParentDefault, Path: "values.yaml", Priority: 1},
		},
	}

	result := Analyze("test", layers)

	if result.Entries[0].Key != "alpha" {
		t.Errorf("expected first key alpha, got %s", result.Entries[0].Key)
	}
	if result.Entries[1].Key != "middle" {
		t.Errorf("expected second key middle, got %s", result.Entries[1].Key)
	}
	if result.Entries[2].Key != "zebra" {
		t.Errorf("expected third key zebra, got %s", result.Entries[2].Key)
	}
}
