# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Children workflow in tests now always start successfully, before their mock's return value is collected/func is run.

## [v0.19.0] - 2022-01-05
### Added
- Added JWT Authorization Provider. This change includes a dependency that uses v2+ go modules. They no longer match import paths, meaning that we have to **drop support for dep & glide** in order to use this. [#1116](https://github.com/uber-go/cadence-client/pull/1116)
### Changed
- Generated proto type were moved out to [cadence-idl](https://github.com/uber/cadence-idl) repository. This is **BREAKING** if you were using `compatibility` package. In that case you will need to update import path from `go.uber.org/cadence/.gen/proto/api/v1` to `github.com/uber/cadence-idl/go/proto/api/v1` [#1138](https://github.com/uber-go/cadence-client/pull/1138)
### Documentation
- Documentation improvements for `client.SignalWorkflow` [#1151](https://github.com/uber-go/cadence-client/pull/1151)


## [v0.18.5] - 2021-11-09
