# VCSInfo

[![License](https://img.shields.io/github/license/jayclassless/vcsinfo.svg?style=flat)](https://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/jayclassless/vcsinfo.svg?branch=master)](https://travis-ci.org/jayclassless/vcsinfo)
[![Coverage Status](https://coveralls.io/repos/github/jayclassless/vcsinfo/badge.svg?branch=master)](https://coveralls.io/github/jayclassless/vcsinfo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jayclassless/vcsinfo)](https://goreportcard.com/report/github.com/jayclassless/vcsinfo)

## Overview

Inspired by [vcprompt](https://bitbucket.org/gward/vcprompt) and its [Python
 port](https://github.com/djl/vcprompt), VCSInfo is a tool that allows you to
extract information from a local version control repository directory and
generate a string that you can embed in your terminal prompt. For example, in
Bash, you could do something like this:

    $ export PS1="\u@\h:\w \$(vcsinfo)\$ "

And end up with a prompt that looks like:

    yourname@yourhost:~ $ cd mycoolproject
    yourname@yourhost:~/mycoolproject git[master]$ touch somefile.txt
    yourname@yourhost:~/mycoolproject git[master+]$ cd ..
    yourname@yourhost:~ $

It's rather handy for knowing the state of the repository you're working in
without the need to always manually invoke one or more commands depending on
the VCS you're using.

Why recreate something that already exists? A few reasons: niether version of
vcprompt has been updated in several years, the Python version could be slow,
the C version was missing some VCS systems, and I wanted an excuse to learn Go,
which is what this implementation was written in.


## Installation

VCSInfo is a simple program -- it's one executable that you should put in your
path somewhere. There are several ways to get your hands on it:

* Download from our [GitHub
  Releases](https://github.com/jayclassless/vcsinfo/releases). Every release
  will be available here with pre-built binaries for all the platforms we
  support. We'll also provide RPMs, DEBs, and SNAPs.

* Homebrew. We provide a custom
  [tap](https://github.com/jayclassless/homebrew-vcsinfo) that allows Homebrew
  users to easily install vcsinfo.

      $ brew tap jayclassless/vcsinfo
      $ brew install vcsinfo

* Compile from [source](https://github.com/jayclassless/vcsinfo). If you're
  comfortable build Go projects, you're welcome to retrieve the source and
  build it yourself.


## Usage

TBD

