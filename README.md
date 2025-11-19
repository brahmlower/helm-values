
# Helm Schema

A helm plugin for generating schema for chart values.

## Schema Generation

```
helm values schema
```

or

```
helm values schema ./charts/my-chart/values.yaml
```

Options:

```
Generate values schema

Usage:
  helm-values schema [flags]

Flags:
      --chart-dir string     path to the chart directory
      --dry-run              don't write changes to disk
  -h, --help                 help for schema
      --log-level string     log level (debug, info, warn, error, fatal, panic) (default "warn")
      --schema-file string   path to the schema-file file (default "values.schema.json")
      --stdout               write to stdout
      --strict               fail on doc comment parsing errors
```

## Docs Generation

```
helm spec docs
```

or

```
helm spec docs ./my-chart.values.yaml
```

Options:
```
Generate values docs

Usage:
  helm-values docs [flags]

Flags:
  -h, --help                 help for docs
```
