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

> **Helm v4 users:** Helm v4 currently requires `--verify=false` for plugin installs from GitHub. This is a [known Helm v4 issue](https://github.com/helm/helm/issues/31490) affecting all plugins, not specific to helm-blame:
> ```bash
> helm plugin install https://github.com/Bisman-Singh/helm-blame --verify=false
> ```

Other install methods:

```bash
# Go install
go install github.com/Bisman-Singh/helm-blame/cmd/helm-blame@latest

# Docker
docker run --rm ghcr.io/bisman-singh/helm-blame:latest --help
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

## Correctness

helm-blame re-implements Helm's values merge logic by reading YAML files directly rather than depending on the Helm SDK. This keeps the binary small and the dependency tree minimal, but it means our merge behavior must match Helm's exactly.

We've validated against three real-world charts:

| Chart | Entries | Subcharts |
|-------|---------|-----------|
| bitnami/nginx | 196 | 0 |
| kube-prometheus-stack | 1,493 | 5 |
| gitlab/gitlab | 2,783 | 15 |

Tested against Helm v4.0.4. If you find a value where helm-blame's provenance or final answer differs from what `helm template` produces, **that's a bug** — please [file an issue](https://github.com/Bisman-Singh/helm-blame/issues).

## Known Limitations

Edge cases we handle correctly:
- Null values delete the entire subtree (children are pruned)
- Lists are replaced entirely, not merged element-by-element
- `--set tags={a,b,c}` is parsed as an array
- `--set key=null` produces a nil value that deletes the subtree
- Escaped dots in `--set` (`kubernetes\.io/os`) work

Edge cases not yet handled:

**Global value propagation.** Helm propagates `.Values.global.*` into every subchart automatically. helm-blame shows `global.domain` and `sub.global.domain` as separate keys rather than showing that the parent global overrides the subchart global. Planned for v0.2.

**Subchart aliases.** Charts can define `dependencies[].alias` in Chart.yaml to rename a subchart. helm-blame doesn't resolve aliases — it uses the directory name. Planned for v0.3.

**Conditional dependencies.** Charts can use `condition: foo.enabled` to skip a dependency entirely. helm-blame doesn't evaluate conditions — it always loads subchart values. Planned for v0.3.

## Roadmap

- Cluster-aware mode: `helm blame <release> -n <namespace>` to trace values of a live release
- Diff mode: compare local values chain against a running release
- Pre-commit hook to block PRs with shadowed values
- ArgoCD / Flux `valuesFrom` support
- VS Code extension with inline hover

## License

MIT
