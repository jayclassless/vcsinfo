# VCSInfo Change Log

All notable changes to this project will be documented in this file. As best we
can, this project will adhere to [Semantic Versioning](https://semver.org).


## [0.2.0] - 2019-02-10

### Added

* Added the ability to indicate the presence of staged/shelved changes.
* Will now bail on retreiving VCS information if a file named ``.novcsinfo`` is
  found in the directory or a parent directory while searching for the
  repository root.
* Added command-line options that can be used to cover the same functionality
  as all the environment variables that are used.
* Added options to specify per-VCS format strings.

### Fixed

* Fixed an issue where a crash would occur if executed somewhere within a
  repository's ``.git`` directory.


## [0.1.0] - 2019-02-05

* Initial public release.

