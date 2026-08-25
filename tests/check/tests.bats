#!/usr/bin/env bats

setup() {
    if command -v brew &> /dev/null; then
        BREW_PREFIX="$(brew --prefix)"
        load "${BREW_PREFIX}/lib/bats-support/load.bash"
        load "${BREW_PREFIX}/lib/bats-assert/load.bash"
    else
        NPM_PREFIX="$(npm config get prefix)"
        load "${NPM_PREFIX}/lib/node_modules/bats-support/load.bash"
        load "${NPM_PREFIX}/lib/node_modules/bats-assert/load.bash"
    fi

    cat > Chart.yaml <<'EOF'
apiVersion: v2
name: check-test-chart
description: A chart for exercising --check
version: 0.1.0
annotations:
  values-schema: https://example.com/refs/tags/helm-chart-0.1.0/values.schema.json
EOF

    cat > values.yaml <<'EOF'
foo: bar
EOF
}

teardown() {
    rm -f Chart.yaml values.yaml values.schema.json README.md
}

@test "schema --check fails before the schema has been generated" {
    run helm values schema . --check
    assert_failure
    assert_output --partial "values.schema.json is stale"
}

@test "schema --check passes once the schema is up to date" {
    helm values schema . --write-modeline=false

    run helm values schema . --check
    assert_success
}

@test "schema --check fails when values.yaml changes without regenerating" {
    helm values schema . --write-modeline=false
    echo "baz: qux" >> values.yaml

    run helm values schema . --check
    assert_failure
    assert_output --partial "values.schema.json is stale"
}

@test "schema --check fails when the values-schema annotation doesn't reference the chart version" {
    helm values schema . --write-modeline=false
    sed -i.bak 's/version: 0.1.0/version: 0.2.0/' Chart.yaml
    rm -f Chart.yaml.bak

    run helm values schema . --check
    assert_failure
    assert_output --partial "does not reference current chart version"
}

@test "docs --check fails before the docs have been generated" {
    run helm values docs . --check
    assert_failure
    assert_output --partial "README.md is stale"
}

@test "docs --check passes once the docs are up to date" {
    helm values docs .

    run helm values docs . --check
    assert_success
}

@test "docs --check fails when values.yaml changes without regenerating" {
    helm values docs .
    echo "baz: qux" >> values.yaml

    run helm values docs . --check
    assert_failure
    assert_output --partial "README.md is stale"
}
