FROM golang:1.13

RUN apt-get update && \
    apt-get install --yes --no-install-recommends --verbose-versions \
        git \
        mercurial \
        subversion \
        cvs \
        darcs \
        fossil \
        bzr && \
    rm -rf /var/lib/apt/lists/*

ENV USER fake

