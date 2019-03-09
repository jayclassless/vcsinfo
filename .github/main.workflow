workflow "Test & Release" {
    on = "push"
    resolves = ["release"]
}

action "release" {
    needs = ["release-filter"]
    uses = "docker://goreleaser/goreleaser:v0.102"
    args = "release"
    secrets = ["GITHUB_TOKEN"]
}

action "release-filter" {
    needs = ["lint"]
    uses = "actions/bin/filter@master"
    args = "tag"
}

action "lint" {
    needs = ["test"]
    uses = "./.github/actions/testenv"
    args = "make lint"
}

action "test" {
    uses = "./.github/actions/testenv"
    args = "make ci-gha"
}

action "build" {
    uses = "docker://goreleaser/goreleaser:v0.102"
    args = "--snapshot --rm-dist"
    secrets = ["SNAPSHOT_VERSION"]
}

