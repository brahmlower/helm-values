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

    touch values.yaml
    task run -- pre-commit
}

teardown() {
    rm -f values.yaml
    rm -f .pre-commit-config.yaml
}

@test "Generate helm values schema" {
    pre-commit run --files values.yaml
    pre-commit run --files values.yaml | grep 'Generate Helm values schema' | grep 'Passed'
    assert_success
}

@test "Generate helm values documentation" {
    pre-commit run --files values.yaml | grep 'Generate Helm values documentation' | grep Passed
    assert_success
}
