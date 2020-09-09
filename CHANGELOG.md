# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Username formatting now correctly only prints FullUsername field ([#126](https://github.com/cyberark/conjur-authn-k8s-client/issues/126))

### Changed
- Wait slightly for the client certificate file to exist after login before
  raising an error.
  [cyberark/conjur-authn-k8s-client#119](https://github.com/cyberark/conjur-authn-k8s-client/issues/119)

## [0.18.0] - 2020-04-21
### Added
- Design for making project FIPS compliant to support users that require it -
  [design](design/fips-compliance.md), [cyberark/conjur-authn-k8s-client#106](https://github.com/cyberark/conjur-authn-k8s-client/issues/106)

### Changed
- The project now uses `goboring/golang` as its base image to be FIPS compliant
  [cyberark/conjur-authn-k8s-client#113](https://github.com/cyberark/conjur-authn-k8s-client/issues/113)
- The authenticator-client now runs as a limited user in the Docker image
  instead of as root, which is best practice and better follows the principle of
  least privilege 
  ([cyberark/conjur-authn-k8s-client#111](https://github.com/cyberark/conjur-authn-k8s-client/pull/111))

## [0.17.0] - 2020-04-07
### Added
- Authenticator client prints its version upon startup (#93)

## [0.16.1] - 2020-02-18
### Fixed
- Only publish to DockerHub / RH registry when there is a new version (#72, #74,
  #79, #83)

### Changed
- Clean up implementation of default CONJUR_VERSION and add unit tests (#80)

### Added
- Added pipeline step to validate CHANGELOG format and update CHANGELOG to meet
  keepachangelog standard (#82)

## [0.16.0] - 2020-01-21
### Changed
- Enable authenticating hosts that have their application identity defined in
  annotations instead of in the id. Hosts that have their application identity
  in the id can be authenticated as well.

## [0.15.0] - 2019-11-26
### Changed
- Enable authenticating hosts that are defined anywhere in the policy tree, instead
  of only hosts that are defined under `conjur/authn-k8s/<service-id>/apps`.

## [0.14.0] - 2019-09-04
### Added
- Added a `log` package with a centralized file for log messages
- Added a constructor for `Authenticator` that receives an AccessToken

### Changed
- Moved all AccessToken related work to a separate package
- Moved all log related work to the `log` package
- NewFromEnv **signature has changed** - method does not take input parameters 
  anymore and is using default values for `tokenFilePath` & `clientCertPath`.
  These parameters can also be set as environment variables:
    - `tokenFilePath` can be set with `CONJUR_AUTHN_TOKEN_FILE`
    - `clientCertPath` can be set with `CONJUR_CLIENT_CERT_PATH`

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

[Unreleased]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.18.0...HEAD
[0.18.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.16.1...v0.17.0
[0.16.1]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.16.0...v0.16.1
[0.16.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.11.1...v0.12.0
[0.11.1]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.10.2...v0.11.0
[0.10.2]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.10.1...v0.10.2
