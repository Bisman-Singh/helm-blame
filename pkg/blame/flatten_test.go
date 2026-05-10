package blame

import (
	"testing"
)

func TestFlattenSimple(t *testing.T) {
	values := map[string]interface{}{
		"replicaCount": 3,
		"image": map[string]interface{}{
			"repository": "nginx",
			"tag":        "1.25",
		},
	}

	result := FlattenValues(values)

	expected := map[string]interface{}{
		"image.repository": "nginx",
		"image.tag":        "1.25",
		"replicaCount":     3,
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(result))
	}

	for _, kv := range result {
		exp, ok := expected[kv.Key]
		if !ok {
			t.Errorf("unexpected key: %s", kv.Key)
			continue
		}
		if kv.Value != exp {
			t.Errorf("key %s: expected %v, got %v", kv.Key, exp, kv.Value)
		}
	}
}

func TestFlattenNested(t *testing.T) {
	values := map[string]interface{}{
		"ingress": map[string]interface{}{
			"enabled": true,
			"annotations": map[string]interface{}{
				"kubernetes.io/ingress.class": "nginx",
			},
		},
	}

	result := FlattenValues(values)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}

	lookup := make(map[string]interface{})
	for _, kv := range result {
		lookup[kv.Key] = kv.Value
	}

	if lookup["ingress.enabled"] != true {
		t.Error("expected ingress.enabled = true")
	}
	if lookup["ingress.annotations.kubernetes.io/ingress.class"] != "nginx" {
		t.Error("expected ingress annotation value")
	}
}

func TestFlattenArray(t *testing.T) {
	values := map[string]interface{}{
		"hosts": []interface{}{
			map[string]interface{}{
				"name": "example.com",
				"port": 443,
			},
			map[string]interface{}{
				"name": "api.example.com",
				"port": 8080,
			},
		},
	}

	result := FlattenValues(values)

	lookup := make(map[string]interface{})
	for _, kv := range result {
		lookup[kv.Key] = kv.Value
	}

	if lookup["hosts[0].name"] != "example.com" {
		t.Errorf("expected hosts[0].name = example.com, got %v", lookup["hosts[0].name"])
	}
	if lookup["hosts[1].port"] != 8080 {
		t.Errorf("expected hosts[1].port = 8080, got %v", lookup["hosts[1].port"])
	}
}

func TestFlattenEmpty(t *testing.T) {
	result := FlattenValues(map[string]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestFlattenSorted(t *testing.T) {
	values := map[string]interface{}{
		"zebra": "z",
		"alpha": "a",
		"middle": map[string]interface{}{
			"beta": "b",
		},
	}

	result := FlattenValues(values)

	if result[0].Key != "alpha" {
		t.Errorf("expected first key = alpha, got %s", result[0].Key)
	}
	if result[1].Key != "middle.beta" {
		t.Errorf("expected second key = middle.beta, got %s", result[1].Key)
	}
	if result[2].Key != "zebra" {
		t.Errorf("expected third key = zebra, got %s", result[2].Key)
	}
}
