FROM golang:1.12

RUN apt-get update && \
    apt-get install --yes --no-install-recommends --verbose-versions \
        git \
        mercurial \
        subversion \
        cvs \
        darcs \
        fossil \
        bzr && \
    rm -rf /var/lib/apt/lists/* && \
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.15.0

ENV USER fake

