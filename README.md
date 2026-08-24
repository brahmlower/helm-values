
# Helm Values

A helm plugin for generating schema and docs for chart values.

[![Release](https://img.shields.io/github/v/release/brahmlower/helm-values.svg?logo=github)](https://github.com/brahmlower/helm-values/releases)
[![Tests](https://github.com/brahmlower/helm-values/actions/workflows/tests.yaml/badge.svg)](https://github.com/brahmlower/helm-values/actions/workflows/tests.yaml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Updating](#updating)
  - [Uninstalling](#uninstalling)
  - [Quickstart](#quickstart)
- [Usage](#usage)
  - [Generate Schema](#generate-schema)
  - [Generate Docs](#generate-docs)
  - [Schema Comments](#schema-comments)
  - [Docs Template API](#docs-templating-api)
    - [Built-In Templates](#built-in-templates)
    - [Extra Templates](#extra-templates)
    - [Template Context](#template-context)
    - [Sprig Functions](#sprig-functions)
    - [Additional Functions](#additional-functions)
- [Development Roadmap](#development-roadmap)
- [Development Notes](#development-notes)

## Getting Started

### Installation

Install the plugin:

```
helm plugin install https://github.com/brahmlower/helm-values/releases/download/0.2.0/values-0.2.0.tgz
```

Releases are signed, and ship with a `.tgz.prov` file alongside the plugin tarball. To verify
the signature on install, fetch the maintainer's public key from [`KEYS.asc`](./KEYS.asc) and
pass it as the verification keyring:

```
curl -sL https://raw.githubusercontent.com/brahmlower/helm-values/main/KEYS.asc | gpg --dearmor > helm-values-keys.gpg
helm plugin install https://github.com/brahmlower/helm-values/releases/download/0.2.0/values-0.2.0.tgz --keyring ./helm-values-keys.gpg
```

Without `--keyring`, `helm plugin install` looks for verification keys in `~/.gnupg/pubring.gpg`,
which modern GnuPG versions don't populate by default — so verification will fail unless you
either use `--keyring` as shown above or pass `--verify=false` to skip verification.

Add pre-commit hooks (optional):

```
helm values pre-commit
```

### Updating
Update the plugin using its own builtin update command:

```
helm values update
```

This simply uninstalls the existing version and installs the latest version from github.

> [!NOTE]
> Since releases are signed, `helm values update` currently fails signature verification
> (it doesn't have a `--keyring` to check against). Until that's addressed, update manually
> using the verified or `--verify=false` install commands above instead.

Helm's builtin update mechanism for plugins only works with VCS-based installations.
This plugin uses tarball installation for simplicity, meaning the builtin update mechanism won't work.

Using the builtin update mechanism will give the following error:
> cannot get information about plugin source

### Uninstalling

```
helm plugin uninstall values
```

### Quickstart

Manually generate your values schema:

```
helm values schema ./path/to/my/chart
```

Manually generate your values docs:

```
helm values docs ./path/to/my/chart
```

## Usage

### Generate Schema

Options:

```
Generate values schema

Usage:
  helm-values schema [flags] chart_dir [...chart_dir]

Flags:
      --dry-run            don't write changes to disk
  -h, --help               help for schema
      --log-level string   log level (debug, info, warn, error, fatal, panic) (default "warn")
      --stdout             write to stdout
      --strict             fail on doc comment parsing errors
      --write-modeline     write modeline to values file (default true)
```

> [!TIP]
> The [redhat.vscode-yaml](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)
> extension (commonly used for validating yaml schema) renders tooltips as plaintext despite
> descriptions containing markdown. (see the github issue [here](https://github.com/redhat-developer/vscode-yaml/issues/454)
> for more details)
>
> As of Dec 2025, the extension will render the `markdownDescription` property as markdown. The following
> `jq` line can be run to duplicate the description as a `markdownDescription` for optimal readability.
>
> ```
> jq 'walk(if type == "object" and .description then . = . * {"markdownDescription": .description} else . end)' ./path/to/schema.values.yaml
> ```

### Generate Docs

Options:

```
Generate values docs

Usage:
  helm-values docs [flags] chart_dir [...chart_dir]

Flags:
      --dry-run                  don't write changes to disk
      --extra-templates string   glob path to extra templates
  -h, --help                     help for docs
      --log-level string         log level (debug, info, warn, error, fatal, panic) (default "warn")
      --markup string            markup language (md, markdown, rst, restructuredtext)
      --output string            path to output (defaults to README.md or README.rst based on markup)
      --stdout                   write to stdout
      --strict                   fail on doc comment parsing errors
      --template string          path to template (defaults to README.md.tmpl or README.rst.tmpl based on markup)
      --use-default              uses default template unless a custom template is present (default true)
```

### Values Modeline

The modeline subcommand is useful for setting a chart's schema in the yaml modeline of a values file.

For instance, running the following:

```
helm values modeline brahmlower-kiwix/kiwix ./kiwix-values.yaml --version 0.1.1
```

results in the following line being added to the top of the document:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/brahmlower/helm-kiwix/refs/tags/kiwix-0.1.1/charts/kiwix/values.schema.json

...
```

Options:

```
Add yaml-language-server modeline to values file

Usage:
  helm-values modeline [flags] chart_ref [values_file]

Flags:
  -f, --force              replace existing modeline
  -h, --help               help for modeline
      --log-level string   log level (debug, info, warn, error, fatal, panic) (default "warn")
  -p, --parents            create parent directories if they don't exist
      --version string     chart version (for remote charts)
```

### Schema Comments

This plugin simplifies schema markup in the values.yaml comments.

Comments are used as the field description by default. Multiline values are supported.

```yaml
# The foo configuration for my app.
foo: qux
```

<details>
<summary>Resulting jsonschema:</summary>

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "foo": {
      "type": "string",
      "title": "foo",
      "description": "The foo configuration for my app",
      "default": "qux"
    },
  }
}
```
</details><br>

If the header comment is parsable as a yaml object, it will be treated as the schema configuration.

```yaml
# type: string
# minLength: 3
# maxLength: 5
# examples:
#   - foo
#   - bar
#   - bax
# description: The foo configuration for my app.
foo: qux
```

<details>
<summary>Resulting jsonschema:</summary>

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "foo": {
      "type": "string",
      "title": "foo",
      "minLength": 3,
      "maxLength": 5,
      "examples": ["foo", "bar", "baz"],
      "description": "The foo configuration for my app",
      "default": "qux"
    },
  }
}
```
</details><br>

Within the header comment, the description can be provided in a second yaml document for improved readability. This is especially helpful for detailed descriptions.

```yaml
# type: string
# minLength: 3
# maxLength: 5
# examples: [foo, bar, baz]
# ---
# The foo configuration for my app.
#
# Only allows [metasyntactic variable][1] names up to length 5 (excluding quuux, etc).
# Used for XYZ purposes in this fictionalized app.
#
# [1]: https://en.wikipedia.org/wiki/Metasyntactic_variable "metasyntactic variable"
foo: qux
```

<details>
<summary>Resulting jsonschema:</summary>

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "foo": {
      "type": "string",
      "title": "foo",
      "minLength": 3,
      "maxLength": 5,
      "examples": ["foo", "bar", "baz"],
      "description": "The foo configuration for my app.\n\nOnly allows [metasyntactic variable][1] names up to length 5 (excluding quuux, etc).\nUsed for XYZ purposes in this fictionalized app.\n\n[1]: https://en.wikipedia.org/wiki/Metasyntactic_variable \"metasyntactic variable\"",
      "default": "qux"
    },
  }
}
```
</details><br>

The `$ref` and `$schema` properties work too, however any other jsonschema properties will be ignored (including descriptions):

```yaml
# $ref: https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/v1.34.0/_definitions.json#/definitions/io.k8s.api.core.v1.ResourceRequirements
# ---
# Container resources only, recommended 1tb mem, 1,000,000 cpu
resources: {}
```

<details>
<summary>Resulting jsonschema:</summary>

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "resources": {
      "title": "resources",
      "$ref": "https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/v1.34.0/_definitions.json#/definitions/io.k8s.api.core.v1.ResourceRequirements"
    },
  }
}
```
</details><br>


### Docs Templating API

Markdown and ReStructuredText are supported.

#### Built-In Templates

Built-in template names are prefixed with the markup language they support (eg: `md`, `rst`) and are provided the full [TemplateContext](#template-context) for flexibility when being overwritten (see [extra templates](#extra-templates)).

> [!NOTE]
> Parity between markup languages is best effort, but is not guaranteed.

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

#### Extra Templates

Built-in templates can be overwritten by including extra template files!

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

#### Template Context

> [!IMPORTANT]
> This project is under very active development. These are likely to change at any point.

The `TemplateContext` and related sub-structures are defined as follows:

```go
type TemplateContext struct {
	Raw         *RawContext
	ValuesTable []ValuesRow
}

type RawContext struct {
	Chart  *charts.Chart
	Values *jsonschema.Schema
}

type ChartDetails struct {
	Name        string
	Description string
}

type ValuesRow struct {
	Key         string
	Type        string
	Default     string
	Description string
}
```

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

## Development Roadmap

Features inspired by [helm-schema](https://github.com/dadav/helm-schema)
and [helm-docs](https://github.com/norwoodj/helm-docs).

- 0.0.2
  - [ ] Schema Generation
    - [x] Check/validate values file
    - [x] Write to non-default location
    - [x] Write to stdout
    - [x] Update values file with yaml-schema modeline
    - [x] Requirement: helm lint checks
    - [x] Warn on undocumented values property
  - [ ] Docs Generation
    - [x] Mardown & ReStructured Text support
    - [x] Render custom and builtin templates
    - [ ] Support rich template customization
      - [x] Sprig functions
    - [x] "UseDefault=false" flag to require existing gotmpl
  - [x] Helm v4 Plugin support
- 0.1.0
  - [ ] Schema Generation
    - [x] Set examples from comments
    - [ ] ~~Warn on ignored jsonschema property (in cases of $ref/$schema usage)~~
  - [ ] Docs Generation
    - [ ] ~~Template: Table of Contents~~
    - [ ] ~~Helpers for table generation~~
    - [x] Support values order (preserved, alphabetical)
  - [x] fixed bug with null values
  - [x] fixed comment parsing with empty lines
  - [x] fixed values rows not being in a consistent order
  - [x] package organization overhaul
- 0.2.0
  - [x] Built-in plugin update mechanism
- 0.3.0
  - [x] Pre-Commit Hook support
  - [x] Values modeline support
- 0.4.0
  - [ ] Schema Generation
    - [ ] Warn on ignored jsonschema property (in cases of $ref/$schema usage)
    - [ ] Json-Schema Draft 7 support?
    - [ ] Support declaring root level attributes
    - [ ] Objects defined in Definitions sections
  - [ ] Docs Generation
    - [ ] Template: Table of Contents
    - [ ] Helpers for table generation
    - [ ] Support "Deprecated" indicator
    - [ ] Template: Chart Values
      - [x] Values groups
- 0.5.0
  - [ ] Schema Generation
    - [ ] Support declaring and using yaml anchors in doc comments
    - [ ] Root level one-of/any-of/all-of
  - [ ] Docs Generation
    - [ ] TODO: Detect recursive templates
    - [ ] TODO: markdown/rst escaping
- 0.6.0
  - [ ] Template: Chart Dependencies (defined in Chart.yaml)
- some day
  - [ ] validate examples against schema

## Development Notes

Dependencies:
- go
- goreleaser
- taskfile
- helm

Building just the binaries

```
task build:bin
```

Build the binaries and plugin will require providing the `--snapshot` flag if you have any uncommitted git changes.

```
task build:plugin -- --snapshot
```

bats install

```
brew install bats-core
brew tap kaos/shell
brew install bats-assert
```


### New release

Bump the changelog:

```
git cliff --bump minor --output CHANGELOG.md
git add .
git commit -m "release $(git cliff --bumped-version --bump minor)"
git tag -a "$(git cliff --bumped-version --bump minor)" -m ''
```
