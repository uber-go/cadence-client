# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.20.0] - 2022-09-29
### Added
- Added refresh tasks API to client [#1162](https://github.com/uber-go/cadence-client/pull/1162)
- Added go1.18 support [#1164](https://github.com/uber-go/cadence-client/pull/1164)
- Added JitterStart support to client [#1178](https://github.com/uber-go/cadence-client/pull/1178)
- Added dynamic polling feature [1184](https://github.com/uber-go/cadence-client/pull/1184), [#1186](https://github.com/uber-go/cadence-client/pull/1186), [#1192](https://github.com/uber-go/cadence-client/pull/1192)
### Changed
- Stopped retrying on AccessDeniedError, FeatureNotEnabledError [#1167](https://github.com/uber-go/cadence-client/pull/1167)
- Stopped retrying get-workflow-history with an impossibly-short timeout [#1171](https://github.com/uber-go/cadence-client/pull/1171)
- Prevented local-activity panics from taking down the worker process [#1169](https://github.com/uber-go/cadence-client/pull/1169)
- Fixed bug with workflow shadower [#1187](https://github.com/uber-go/cadence-client/pull/1187)
- upgrade thrift to v0.16.0 and tchannel-go to v1.32.1 [#1189](https://github.com/uber-go/cadence-client/pull/1189)

## [v0.19.1] - 2022-06-02
### Changed
- Client will not retry on LimitExceededError [#1170](https://github.com/uber-go/cadence-client/pull/1170)

## [v0.19.0] - 2022-01-05
### Added
- Added JWT Authorization Provider. This change includes a dependency that uses v2+ go modules. They no longer match import paths, meaning that we have to **drop support for dep & glide** in order to use this. [#1116](https://github.com/uber-go/cadence-client/pull/1116)
### Changed
- Generated proto type were moved out to [cadence-idl](https://github.com/uber/cadence-idl) repository. This is **BREAKING** if you were using `compatibility` package. In that case you will need to update import path from `go.uber.org/cadence/.gen/proto/api/v1` to `github.com/uber/cadence-idl/go/proto/api/v1` [#1138](https://github.com/uber-go/cadence-client/pull/1138)
### Documentation
- Documentation improvements for `client.SignalWorkflow` [#1151](https://github.com/uber-go/cadence-client/pull/1151)


## [v0.18.5] - 2021-11-09
