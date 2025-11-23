
# Helm Values

A helm plugin for generating schema and docs for chart values.

[![Tests](https://github.com/brahmlower/helm-values/actions/workflows/tests.yaml/badge.svg)](https://github.com/brahmlower/helm-values/actions/workflows/tests.yaml)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-values)](https://artifacthub.io/packages/search?repo=helm-values)

## TL;DR Getting Started

Install the plugin:

```
helm plugin install https://github.com/brahmlower/helm-values
```

Generate your values shcmea:

```
helm values schema ./path/to/my/chart
```

Generate your values docs:

```
helm values docs ./path/to/my/chart
```


## Schema Generation

```
helm values schema
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

Options:
```
Generate values docs

Usage:
  helm-values docs [flags]

Flags:
      --chart-dir string         path to the chart directory
      --dry-run                  don't write changes to disk
      --extra-templates string   glob path to extra templates
  -h, --help                     help for docs
      --log-level string         log level (debug, info, warn, error, fatal, panic) (default "warn")
      --markup string            markup language (md, markdown, rst, restructuredtext)
      --output string            path to output (defaults to README.md or README.rst based on markup)
      --schema-file string       path to the schema-file file (default "values.schema.json")
      --stdout                   write to stdout
      --strict                   fail on doc comment parsing errors
      --template string          path to template (defaults to README.md.tmpl or README.rst.tmpl based on markup)
      --use-default              uses default template unless a custom template is present (default true)
```

## Template API

Markdown and ReStructuredText are supported.

### Sprig Functions

Functions from [sprig](https://masterminds.github.io/sprig/) version 3.3.0 are available.

### Additional Functions:

These are by no means considered stable, and will almost certainly change before initial stable version.

#### `lpad`

The lpad function adds space to the left until the desired length has been met:

```
lpad "hello" 10
```

The above produces `     hello`

#### `rpad`

The lpad function adds space to the right until the desired length has been met:

```
rpad "hello" 10
```

The above produces `hello     `

#### `maxLen`

The maxLen function returns the largest length in the list of strings:

```
maxLen "hello" "foo" "kubernetes"
```

The above produces `10`

### Template Context

> [!IMPORTANT]
> This project is under very active development. These are likely to change at any point.

The `TemplateContext` and related sub-structures are defined as follows:

```go
type ValuesRow struct {
	Key         string
	Type        string
	Default     string
	Description string
}

type RawContext struct {
	Chart  *charts.Chart
	Values *jsonschema.Schema
}

type TemplateContext struct {
	Raw         *RawContext
	ValuesTable []ValuesRow
}

type ChartDetails struct {
	Name        string
	Description string
}
```

### Template Definitions

Template names are prefixed with the markup language they support. Built-in templates generally take the full [TemplateContext](#template-context) to give maximum flexibility to those who want to [override templates](#overriding-templates).

> [!NOTE]
> Parity between markup languages is maintaned as best as possible, but is not guaranteed.

- `md.header`

  Document title using the chart name declared in Chart.yaml

- `md.description`

  Subtitle description using the description declared in Chart.yaml

- `md.valuesTable`

  Produces a table of values with columns for Key, Type, Default, Description.
  No multiline support.

- `rst.header`

  Document title using the chart name declared in Chart.yaml

- `rst.description`

  Subtitle description using the description declared in Chart.yaml

- `rst.valuesTable`

  Produces a table of values with columns for Key, Type, Default, Description.
  No multiline support.

### Overriding Templates

Built-in templates can be overwritten (in part or in full) by including extra template files!

For example, the default `md.header` template can be overwritten by defining a template with the same name:

```
{{- define "md.header" }}
# {{ .Raw.Chart.Details.Name }} - A chart by me 😎
{{- end }}
```

Now generate the docs and include the extra template file:

```
helm values docs --extra-templates ./readme-helpers.tmpl
```

Docs generation uses the custom template rather than the builtin.

```
$ head -n 2 README.md

# MyChart - A chart by me 😎
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
  - [x] Mardown & ReStructured Text support
  - [x] Render custom and builtin templates
  - [ ] Support rich template customization
    - [x] Sprig functions
    - [ ] Helpers for table generation
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
