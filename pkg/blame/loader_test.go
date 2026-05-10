package blame

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestChart(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Chart.yaml
	writeFile(t, filepath.Join(dir, "Chart.yaml"), `
name: my-app
version: 1.0.0
`)

	// Parent values.yaml
	writeFile(t, filepath.Join(dir, "values.yaml"), `
replicaCount: 1
image:
  repository: nginx
  tag: latest
`)

	// Subchart
	subDir := filepath.Join(dir, "charts", "redis")
	os.MkdirAll(subDir, 0755)
	writeFile(t, filepath.Join(subDir, "values.yaml"), `
port: 6379
maxmemory: 256mb
`)
	writeFile(t, filepath.Join(subDir, "Chart.yaml"), `
name: redis
version: 0.1.0
`)

	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadChartBasic(t *testing.T) {
	dir := createTestChart(t)

	name, layers, err := LoadChart(dir, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if name != "my-app" {
		t.Errorf("expected chart name my-app, got %s", name)
	}

	// Should have subchart default + parent default = 2 layers
	if len(layers) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(layers))
	}

	if layers[0].Source.Type != SourceSubchartDefault {
		t.Errorf("first layer should be subchart default, got %s", layers[0].Source.Type)
	}
	if layers[1].Source.Type != SourceParentDefault {
		t.Errorf("second layer should be parent default, got %s", layers[1].Source.Type)
	}

	// Subchart values should be namespaced under the subchart name
	flat := FlattenValues(layers[0].Values)
	for _, kv := range flat {
		if kv.Key[:6] != "redis." {
			t.Errorf("subchart key should be prefixed with redis., got %s", kv.Key)
		}
	}
}

func TestLoadChartWithValueFiles(t *testing.T) {
	dir := createTestChart(t)

	// Create override files
	prodFile := filepath.Join(t.TempDir(), "production.yaml")
	writeFile(t, prodFile, `
replicaCount: 3
image:
  tag: "1.25"
`)

	name, layers, err := LoadChart(dir, []string{prodFile}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if name != "my-app" {
		t.Errorf("expected my-app, got %s", name)
	}

	// subchart + parent + value file = 3
	if len(layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(layers))
	}

	if layers[2].Source.Type != SourceValueFile {
		t.Errorf("third layer should be value file, got %s", layers[2].Source.Type)
	}
}

func TestLoadChartWithSetFlags(t *testing.T) {
	dir := createTestChart(t)

	_, layers, err := LoadChart(dir, nil, []string{"replicaCount=5", "image.tag=2.0"})
	if err != nil {
		t.Fatal(err)
	}

	// subchart + parent + 2 set flags = 4
	if len(layers) != 4 {
		t.Fatalf("expected 4 layers, got %d", len(layers))
	}

	if layers[2].Source.Type != SourceSetFlag {
		t.Errorf("third layer should be set flag, got %s", layers[2].Source.Type)
	}
	if layers[3].Source.Type != SourceSetFlag {
		t.Errorf("fourth layer should be set flag, got %s", layers[3].Source.Type)
	}
}

func TestLoadChartPriorityOrder(t *testing.T) {
	dir := createTestChart(t)

	prodFile := filepath.Join(t.TempDir(), "prod.yaml")
	writeFile(t, prodFile, `replicaCount: 3`)

	_, layers, err := LoadChart(dir, []string{prodFile}, []string{"replicaCount=5"})
	if err != nil {
		t.Fatal(err)
	}

	// Verify priorities increase
	for i := 1; i < len(layers); i++ {
		if layers[i].Source.Priority < layers[i-1].Source.Priority {
			t.Errorf("layer %d (priority %d) should be >= layer %d (priority %d)",
				i, layers[i].Source.Priority, i-1, layers[i-1].Source.Priority)
		}
	}
}

func TestParseSetFlag(t *testing.T) {
	key, val, err := parseSetFlag("image.tag=1.25")
	if err != nil {
		t.Fatal(err)
	}
	if key != "image.tag" || val != "1.25" {
		t.Errorf("expected image.tag=1.25, got %s=%s", key, val)
	}

	_, _, err = parseSetFlag("invalid")
	if err == nil {
		t.Error("expected error for invalid set flag")
	}
}

func TestBuildNestedMap(t *testing.T) {
	result := buildNestedMap("image.tag", "1.25")

	image, ok := result["image"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested map under image")
	}
	if image["tag"] != "1.25" {
		t.Errorf("expected tag=1.25, got %v", image["tag"])
	}
}

func TestParseSetValueList(t *testing.T) {
	result := parseSetValue("{a,b,c}")
	list, ok := result.([]interface{})
	if !ok {
		t.Fatal("expected list")
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
	if list[0] != "a" || list[1] != "b" || list[2] != "c" {
		t.Errorf("expected [a,b,c], got %v", list)
	}
}

func TestParseSetValueNull(t *testing.T) {
	result := parseSetValue("null")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestParseSetValueString(t *testing.T) {
	result := parseSetValue("hello")
	if result != "hello" {
		t.Errorf("expected hello, got %v", result)
	}
}

func TestLoadChartNoSubcharts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Chart.yaml"), `name: simple`)
	writeFile(t, filepath.Join(dir, "values.yaml"), `port: 8080`)

	_, layers, err := LoadChart(dir, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
}
