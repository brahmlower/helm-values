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

    mkdir -p bad good

    cat > bad/Chart.yaml <<'YAML'
apiVersion: v2
name: bad
description: A chart with an invalid doc comment
version: 0.1.0
YAML

    cat > bad/values.yaml <<'YAML'
# minimum: .inf
bad: 2
YAML

    cat > good/Chart.yaml <<'YAML'
apiVersion: v2
name: good
description: A chart with valid doc comments
version: 0.1.0
YAML

    cat > good/values.yaml <<'YAML'
# -- replicas to run
replicas: 1
YAML
}

teardown() {
    rm -rf bad good
}

@test "schema --strict fails on a bad doc comment without printing usage" {
    run helm values schema bad --strict --write-modeline=false
    assert_failure
    assert_output --partial "doc comment error"
    refute_output --partial "Usage:"
}

@test "schema without --strict only warns on a bad doc comment" {
    run helm values schema bad --write-modeline=false
    assert_success
}

@test "schema --strict still generates the other charts" {
    run helm values schema bad good --strict --write-modeline=false
    assert_failure
    assert_output --partial "schema: bad:"
    assert [ -f good/values.schema.json ]
}

@test "docs --strict fails on a bad doc comment without printing usage" {
    run helm values docs bad --strict
    assert_failure
    assert_output --partial "doc comment error"
    refute_output --partial "Usage:"
}

@test "docs --strict still generates the other charts" {
    run helm values docs bad good --strict
    assert_failure
    assert_output --partial "docs: bad:"
    assert [ -f good/README.md ]
}

@test "schema --check --strict reports check problems alongside errors" {
    run helm values schema --check bad good --strict
    assert_failure
    assert_output --partial "doc comment error"
    assert_output --partial "values.schema.json is stale"
}
