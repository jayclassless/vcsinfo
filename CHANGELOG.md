# VCSInfo Change Log

All notable changes to this project will be documented in this file. As best we
can, this project will adhere to [Semantic Versioning](https://semver.org).


## [0.3.7] - 2021-11-05

### Changed

* Upgraded to Go 1.17


## [0.3.5] - 2020-01-11

### Fixed

* Fixed an issue with Mercurial complaining about untrusted hgrc files.


## [0.3.2] - 2019-09-08

### Fixed

* Fixed an issue with detecting problems in newer versions of Git.

### Changed

* Upgraded to Go 1.13.


## [0.3.1] - 2019-03-03

### Fixed

* Project/Packaging tweaks.


## [0.3.0] - 2019-03-02

### Changed

* Rebuilt with Go 1.12.


## [0.2.1] - 2019-02-22

### Fixed

* Fixed some occasional Mercurial errors.


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

