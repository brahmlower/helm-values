
# Helm Values

A helm plugin for generating schema for chart values.

[![Tests](https://github.com/brahmlower/helm-values/actions/workflows/tests.yaml/badge.svg)](https://github.com/brahmlower/helm-values/actions/workflows/tests.yaml)

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-values)](https://artifacthub.io/packages/search?repo=helm-values)

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

## Development Roadmap

Features inspired by [helm-schema](https://github.com/dadav/helm-schema)
and [helm-docs](https://github.com/norwoodj/helm-docs).

- [ ] Schema Generation
  - [ ] Check/validate values file
  - [x] Write to non-default location
  - [x] Write to stdout
  - [ ] Update values file with yaml-schema reference
  - [ ] Set examples from comments
  - [ ] Json-Schema Draft 6 support?
  - [ ] Json-Schema Draft 7 support?
  - [ ] Support declaring and using yaml anchors in doc comments
  - [ ] Support declaring root level attributes
  - [ ] Root level one-of/any-of/all-of
  - [ ] Requirement: Changes to values.yaml don't violate yamllint checks
  - [ ] Requirement: helm lint checks
- [ ] Docs Generation
  - [ ] Mardown & ReStructured Text support
  - [x] Render custom and builtin templates
  - [ ] Support rich template customization
    - [ ] Sprig functions
  - [ ] Template: Table of Contents
  - [ ] Template: Chart Values
    - [ ] Support "Deprecated" indicator
    - [ ] Values render order
      - [ ] Alphabetical
      - [ ] Preserved
  - [ ] Template: Chart Dependencies (defined in Chart.yaml)
  - [ ] "NoCreateDefault" flag to require existing gotmpl
  - [ ] TODO: markdown/rst escaping
  - [ ] TODO: Detect recursive templates
- [ ] Helm v3 Plugin support (probably won't do)
- [x] Helm v4 Plugin support
- [ ] Pre-Commit Hook support
