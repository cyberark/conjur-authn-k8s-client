# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.22.0] - 2021-09-17
### Added
- Introduces the `conjur-config-cluster-prep.yaml` and `conjur-config-namespace-prep.yaml` raw Kubernetes manifests generated from their corresponding Helm charts. These manifests provide an alternative method of configuring a Kubernetes cluster for the deployment of Conjur-authenticated applications for users unable to use  Helm in their environment.
  [cyberark/conjur-authn-k8s-client#338](https://github.com/cyberark/conjur-authn-k8s-client/issues/338)
- Added user-configurable Helm values for the names of resources created by the `conjur-config-namespace-prep` Helm chart
  [cyberark/conjur-authn-k8s-client#383](https://github.com/cyberark/conjur-authn-k8s-client/issues/383)

### Security
- Upgrades Openssl in Alpine to resolve CVE-2021-3711.
  [cyberark/conjur-authn-k8s-client#392](https://github.com/cyberark/conjur-authn-k8s-client/issues/392)
- Upgrades Alpine to v3.14 to resolve CVE-2021-36159.
  [cyberark/conjur-authn-k8s-client#374](https://github.com/cyberark/conjur-authn-k8s-client/issues/374)

## [0.21.0] - 2021-06-25
### Added
- Introduces the `conjur-config-cluster-prep` and `conjur-config-namespace-prep` Helm charts.
  Together these charts simplify the deployment of Conjur-authenticated applications as part of
  the [Simplified Client Configuration](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/design/simple-client-configuration.md) feature.
  [cyberark/conjur-authn-k8s-client#232](https://github.com/cyberark/conjur-authn-k8s-client/issues/232)
  [cyberark/conjur-authn-k8s-client#249](https://github.com/cyberark/conjur-authn-k8s-client/issues/249)

## [0.20.0] - 2021-06-16
### Fixed
- Fixes bug in error handling within the `VerifyFileExists` method that resulted in a
  panic when the error from `os.Stat` was not `ErrNotExist`. The fix includes introducing
  the `CAKC058` error and log message for a file permissions error and the`CAKC059` error
  and log message for when the path to a file exists but is not a regular file.
  [cyberark/conjur-authn-k8s-client#252](https://github.com/cyberark/conjur-authn-k8s-client/issues/252)

### Changed
- The `CAKC048` log message now shows the release version for release builds
  and no longer includes the git commit hash in the log output.
  [cyberark/conjur-authn-k8s-client#196](https://github.com/cyberark/conjur-authn-k8s-client/issues/196)
- RH base image is now `ubi8/ubi` instead of `rhel7/rhel`.
  [cyberark/conjur-authn-k8s-client#324](https://github.com/cyberark/conjur-authn-k8s-client/pull/324)

## [0.19.1] - 2021-02-08
### Changed
- The `Authenticate` method now parses the authentication response and writes it
  to the token file, without the need to call `ParseAuthenticationResponse`.
  This is a breaking change for software that leverages the
  `github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator`
  Go package (e.g. Secretless and Secrets Provider for Kubernetes); users of the
  Authn-K8s client Docker image are not impacted by this change.
  [cyberark/conjur-authn-k8s-client#180](https://github.com/cyberark/conjur-authn-k8s-client/issues/180)
- The project Golang version is updated from the end-of-life v1.12 to the latest
  version v1.15.
  [cyberark/conjur-authn-k8s-client#206](https://github.com/cyberark/conjur-authn-k8s-client/issues/206)
- The error message raised when the username doesn't include the `host/` prefix
  now suggests that the user check this. Previously the error message did not
  include any information about what was wrong with the username.
  [PR cyberark/conjur-authn-k8s-client#212](https://github.com/cyberark/conjur-authn-k8s-client/pull/212)

### Added
- Support for OpenShift 4.6 was certified as of this release.
- Support for OpenShift 4.7 was certified as of this release.

## [0.19.0] - 2020-10-08
### Added
- Users can set the `DEBUG` environment variable to run the client in debug mode and view more log messages.
  [cyberark/conjur-authn-k8s-client#134](https://github.com/cyberark/conjur-authn-k8s-client/issues/134)
- Errors in the certificate injection process on login are now printed to the client logs.
  [cyberark/conjur-authn-k8s-client#/170](https://github.com/cyberark/conjur-authn-k8s-client/issues/170)

### Changed
- Detailed logs moved from Info to Debug log level to decrease verbosity of log messages.
  [cyberark/conjur-authn-k8s-client#134](https://github.com/cyberark/conjur-authn-k8s-client/issues/134)
- Log level suffix was removed from log identifiers (e.g. `CAKC001**E**` -> `CAKC001`). To
  avoid conflicts, some log identifiers had to be changed. See [log_messages.go](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/pkg/log/log_messages.go)
  for updated log identifiers.
  [cyberark/conjur-authn-k8s-client#134](https://github.com/cyberark/conjur-authn-k8s-client/issues/134)
- Log messages now show microseconds, for clarity and easier troubleshooting.
  [cyberark/conjur-authn-k8s-client#164](https://github.com/cyberark/conjur-authn-k8s-client/issues/164)

## [0.18.1] - 2020-09-13
### Fixed
- Logs now correctly print only the Conjur identity without the policy branch prefix.
  [cyberark/conjur-authn-k8s-client#126](https://github.com/cyberark/conjur-authn-k8s-client/issues/126)
- When authentication fails, the exponential backoff retry is correctly reset so 
  that it will continue to attempt to authenticate until backoff is exhausted.
  [cyberark/conjur-authn-k8s-client#158](https://github.com/cyberark/conjur-authn-k8s-client/issues/158)

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
  [cyberark/conjur-authn-k8s-client#111](https://github.com/cyberark/conjur-authn-k8s-client/pull/111)

## [0.17.0] - 2020-04-07
### Added
- Authenticator client prints its version upon startup (#93)

## [0.16.1] - 2020-02-18
### Fixed
- Only publish to DockerHub / RH registry when there is a new version
  (#72, #74, #79, #83)

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

[Unreleased]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.22.0...HEAD
[0.22.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.19.1...v0.20.0
[0.19.1]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.19.0...v0.19.1
[0.19.0]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.18.1...v0.19.0
[0.18.1]: https://github.com/cyberark/conjur-authn-k8s-client/compare/v0.18.0...v0.18.1
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
[0.10.1]: https://github.com/cyberark/conjur-authn-k8s-client/releases/tag/v0.10.1
