# VCSInfo Change Log

All notable changes to this project will be documented in this file. As best we
can, this project will adhere to [Semantic Versioning](https://semver.org).


## [Unreleased]

### Added

* Added the ability to indicate the presence of staged/shelved changes.
* Will now bail on retreiving VCS information if a file named ``.novcsinfo`` is
  found in the directory or a parent directory while searching for the
  repository root.
* Added command-line options that can be used to cover the same functionality
  as all the environment variables that are used.


## [0.1.0] - 2019-02-05

* Initial public release.

