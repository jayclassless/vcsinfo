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
    needs = ["test", "lint"]
    uses = "actions/bin/filter@master"
    args = "tag"
}

action "test" {
    uses = "./.github/actions/testenv"
    args = "make ci-gha"
}

action "lint" {
    uses = "./.github/actions/testenv"
    args = "make lint"
}

action "build" {
    uses = "docker://goreleaser/goreleaser:v0.102"
    args = "--snapshot --rm-dist"
    secrets = ["SNAPSHOT_VERSION"]
}

