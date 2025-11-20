
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
helm values docs
```

or

```
helm values docs ./my-chart.values.yaml
```

Options:
```
Generate values docs

Usage:
  helm-values docs [flags]

Flags:
      --chart-dir string         path to the chart directory
      --dry-run                  don't write changes to disk
      --extra-templates string   path to extra templates directory
  -h, --help                     help for docs
      --log-level string         log level (debug, info, warn, error, fatal, panic) (default "warn")
      --schema-file string       path to the schema-file file (default "values.schema.json")
      --stdout                   write to stdout
      --strict                   fail on doc comment parsing errors
```
