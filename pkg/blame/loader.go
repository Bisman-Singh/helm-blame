package blame

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	prioritySubchartBase = 0
	priorityParentBase   = 1
	priorityValueFile    = 100
	prioritySetFlag      = 200
)

// LoadChart reads a chart directory and builds the ordered value layers.
// valueFiles are the -f/--values file paths in order (left to right).
// setValues are --set key=value pairs in order.
func LoadChart(chartDir string, valueFiles []string, setValues []string) (string, []Layer, error) {
	chartName, err := readChartName(chartDir)
	if err != nil {
		return "", nil, fmt.Errorf("reading Chart.yaml: %w", err)
	}

	var layers []Layer

	// 1. Subchart defaults (lowest priority)
	subchartLayers, err := loadSubchartDefaults(chartDir)
	if err != nil {
		return "", nil, fmt.Errorf("loading subchart defaults: %w", err)
	}
	layers = append(layers, subchartLayers...)

	// 2. Parent chart defaults
	parentValues, err := loadValuesFile(filepath.Join(chartDir, "values.yaml"))
	if err != nil && !os.IsNotExist(err) {
		return "", nil, fmt.Errorf("loading parent values.yaml: %w", err)
	}
	if parentValues != nil {
		layers = append(layers, Layer{
			Values: parentValues,
			Source: Source{
				Type:     SourceParentDefault,
				Path:     "values.yaml",
				Priority: priorityParentBase,
			},
		})
	}

	// 3. Value files (-f flags, left to right)
	for i, f := range valueFiles {
		vals, err := loadValuesFile(f)
		if err != nil {
			return "", nil, fmt.Errorf("loading values file %s: %w", f, err)
		}
		if vals != nil {
			layers = append(layers, Layer{
				Values: vals,
				Source: Source{
					Type:     SourceValueFile,
					Path:     filepath.Base(f),
					Priority: priorityValueFile + i,
				},
			})
		}
	}

	// 4. --set flags (highest priority)
	for i, s := range setValues {
		key, value, err := parseSetFlag(s)
		if err != nil {
			return "", nil, fmt.Errorf("parsing --set %s: %w", s, err)
		}
		vals := buildNestedMap(key, value)
		layers = append(layers, Layer{
			Values: vals,
			Source: Source{
				Type:     SourceSetFlag,
				Path:     fmt.Sprintf("--set %s", s),
				Priority: prioritySetFlag + i,
			},
		})
	}

	return chartName, layers, nil
}

// readChartName extracts the chart name from Chart.yaml.
func readChartName(chartDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	if err != nil {
		return "", err
	}
	var chart struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return "", err
	}
	if chart.Name == "" {
		return filepath.Base(chartDir), nil
	}
	return chart.Name, nil
}

// loadValuesFile reads and parses a YAML values file.
func loadValuesFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return values, nil
}

// loadSubchartDefaults reads values.yaml from each subchart in charts/.
func loadSubchartDefaults(chartDir string) ([]Layer, error) {
	chartsDir := filepath.Join(chartDir, "charts")
	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var layers []Layer
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subValuesPath := filepath.Join(chartsDir, entry.Name(), "values.yaml")
		vals, err := loadValuesFile(subValuesPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if vals != nil {
			// Namespace subchart values under the subchart name,
			// matching how Helm merges parent values into subcharts.
			// e.g., grafana subchart's "adminUser" becomes "grafana.adminUser"
			namespacedVals := map[string]interface{}{
				entry.Name(): vals,
			}
			layers = append(layers, Layer{
				Values: namespacedVals,
				Source: Source{
					Type:     SourceSubchartDefault,
					Path:     fmt.Sprintf("charts/%s/values.yaml", entry.Name()),
					Priority: prioritySubchartBase,
				},
			})
		}
	}
	return layers, nil
}

// parseSetFlag splits "key=value" into key and value.
func parseSetFlag(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected key=value")
	}
	return parts[0], parts[1], nil
}

// buildNestedMap creates a nested map from a dot-separated key path.
// Example: "image.tag" with value "1.25" → {"image": {"tag": "1.25"}}
func buildNestedMap(key string, value string) map[string]interface{} {
	parts := strings.Split(key, ".")
	result := make(map[string]interface{})
	current := result

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			next := make(map[string]interface{})
			current[part] = next
			current = next
		}
	}

	return result
}
