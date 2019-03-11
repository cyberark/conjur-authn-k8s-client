# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.13.0] - 2019-03-08

### Fixed
- Fixed issues with certificate expiration not being handled properly

### Added
- Added ability to specify token timeout with `CONJUR_TOKEN_TIMEOUT` variable

### Changed
- Modules have been reorganized to DRY out the main runner module

## [0.12.0] - 0000-00-00

### Changed
- Reorganized file structure of project to make importable

## [0.11.1] - 0000-00-00

### Fixed
- Fixed bug with request body during v4 authentication.

## [0.11.0] - 0000-00-00

### Added
- Added support for Conjur v5.
- Added `CONJUR_VERSION` env variable ('4' or '5', defaults to '5').

## [0.10.2] - 0000-00-00

### Added
- Added a RedHat-certified version of the image.

## [0.10.1] - 0000-00-00

### Fixed
- Fix an issue where sidecar fails when not run as root user.

[Unreleased]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.13.0...HEAD
[0.13.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.12.0...v0.13.0
