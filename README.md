# helm-blame

Trace where every Helm value comes from.

For each leaf value in a chart's merged values tree, `helm blame` shows whether it came from a subchart default, the parent chart's `values.yaml`, an override file (`-f`), or a `--set` flag.

```
$ helm blame ./my-chart -f production.yaml --set replicaCount=5

KEY               VALUE            SOURCE
---               -----            ------
image.pullPolicy  IfNotPresent     values.yaml (chart default)
image.repository  nginx            values.yaml (chart default)
image.tag         1.25.3           production.yaml (override file)
ingress.enabled   true             production.yaml (override file)
ingress.host      app.example.com  production.yaml (override file)
redis.port        6379             charts/redis/values.yaml (subchart default)
replicaCount      5                --set replicaCount=5 (CLI override)
```

## Why

Helm merges values from multiple sources in a specific order. When something is wrong, you need to know: which source set this value, and what did it override?

Today that means opening every values file, mentally tracing the merge order, and hoping you get it right. `helm blame` does it in one command.

## Install

```bash
helm plugin install https://github.com/Bisman-Singh/helm-blame
```

Or build from source:

```bash
git clone https://github.com/Bisman-Singh/helm-blame.git
cd helm-blame
make build
./bin/helm-blame --help
```

## Usage

```bash
# Analyze a chart with its defaults
helm blame ./my-chart

# With override files (left to right, rightmost wins)
helm blame ./my-chart -f base.yaml -f production.yaml

# With --set flags
helm blame ./my-chart --set replicaCount=5 --set image.tag=2.0

# Combined
helm blame ./my-chart -f base.yaml -f staging.yaml --set replicaCount=3

# Show shadowed values (overridden by higher-priority sources)
helm blame ./my-chart -f production.yaml --set replicaCount=5 --show-shadowed

# JSON output
helm blame ./my-chart -f production.yaml -o json
```

## Shadowed Values

Use `--show-shadowed` to see which sources were overridden:

```
KEY           VALUE  SOURCE
---           -----  ------
replicaCount  5      --set replicaCount=5 (CLI override) *
                       ← shadowed: values.yaml (chart default)
                       ← shadowed: production.yaml (override file)
```

This tells you that `replicaCount` was set in three places, and the `--set` flag won.

## Merge Precedence

Follows Helm's documented merge order (lowest to highest priority):

1. Subchart `values.yaml` defaults
2. Parent chart `values.yaml` defaults
3. `-f` / `--values` files (left to right, rightmost wins)
4. `--set` / `--set-string` flags (left to right, rightmost wins)

## Output Formats

**Table** (default): human-readable columns with optional shadowed display.

**JSON** (`-o json`): machine-readable output for CI pipelines and policy gates.

```json
{
  "chart": "my-app",
  "count": 8,
  "entries": [
    {
      "key": "replicaCount",
      "value": "5",
      "source": {
        "type": "CLI override",
        "path": "--set replicaCount=5",
        "priority": 200
      },
      "shadowed": [
        {"type": "chart default", "path": "values.yaml", "priority": 1},
        {"type": "override file", "path": "production.yaml", "priority": 100}
      ]
    }
  ]
}
```

## Roadmap

- Cluster-aware mode: `helm blame <release> -n <namespace>` to trace values of a live release
- Diff mode: compare local values chain against a running release
- Pre-commit hook to block PRs with shadowed values
- ArgoCD / Flux `valuesFrom` support
- VS Code extension with inline hover

## License

MIT
