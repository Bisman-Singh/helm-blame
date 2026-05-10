package main

import (
	"fmt"
	"os"

	"github.com/Bisman-Singh/helm-blame/pkg/blame"
	"github.com/Bisman-Singh/helm-blame/pkg/output"
)

// Set by goreleaser ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const usage = `helm-blame - trace where every Helm value comes from

Usage:
  helm blame <chart-dir> [flags]

Flags:
  -f, --values <file>    Override values file (can be specified multiple times)
  --set <key=value>      Set individual values (can be specified multiple times)
  -o, --output <format>  Output format: table, json (default: table)
  --show-shadowed        Show values that were overridden by higher-priority sources
  -h, --help             Show this help message

Examples:
  helm blame ./my-chart
  helm blame ./my-chart -f production.yaml --set replicaCount=5
  helm blame ./my-chart -f base.yaml -f staging.yaml -o json
  helm blame ./my-chart --show-shadowed`

func main() {
	args := parseArgs(os.Args[1:])

	if args.showVersion {
		fmt.Printf("helm-blame %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	if args.help || args.chartDir == "" {
		fmt.Println(usage)
		os.Exit(0)
	}

	chartName, layers, err := blame.LoadChart(args.chartDir, args.valueFiles, args.setValues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := blame.Analyze(chartName, layers)

	if err := output.Render(os.Stdout, result, args.outputFormat, args.showShadowed); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type cliArgs struct {
	chartDir     string
	valueFiles   []string
	setValues    []string
	outputFormat string
	showShadowed bool
	showVersion  bool
	help         bool
}

func parseArgs(args []string) cliArgs {
	parsed := cliArgs{
		outputFormat: output.FormatTable,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--values":
			if i+1 < len(args) {
				i++
				parsed.valueFiles = append(parsed.valueFiles, args[i])
			}
		case "--set":
			if i+1 < len(args) {
				i++
				parsed.setValues = append(parsed.setValues, args[i])
			}
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				parsed.outputFormat = args[i]
			}
		case "--show-shadowed":
			parsed.showShadowed = true
		case "-h", "--help":
			parsed.help = true
		case "-v", "--version":
			parsed.showVersion = true
		default:
			if parsed.chartDir == "" && args[i][0] != '-' {
				parsed.chartDir = args[i]
			}
		}
	}

	return parsed
}
