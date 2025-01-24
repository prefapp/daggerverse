# Changelog

## [1.3.0](https://github.com/prefapp/daggerverse/compare/hydrate-orchestrator-v1.2.0...hydrate-orchestrator-v1.3.0) (2025-01-24)


### Features

* support firestartr configs, generate  `respositories.yaml` and  `environments.yaml` in runtime ([#96](https://github.com/prefapp/daggerverse/issues/96)) ([79d9cd9](https://github.com/prefapp/daggerverse/commit/79d9cd96cb37637f23751a87aa3c06802f1ad94b))
* support multiple helm credentials with `~/.config/helm/registry/config.json` ([#90](https://github.com/prefapp/daggerverse/issues/90)) ([07733a2](https://github.com/prefapp/daggerverse/commit/07733a2db842a5e79b1c6680db691d22dcde28d2))
* support sys services ([#92](https://github.com/prefapp/daggerverse/issues/92)) ([858c21d](https://github.com/prefapp/daggerverse/commit/858c21d7114ecca78fd0a017daa5df2ed6fe3992))

## [1.2.0](https://github.com/prefapp/daggerverse/compare/hydrate-orchestrator-v1.1.0...hydrate-orchestrator-v1.2.0) (2025-01-22)


### Features

* [state-repo] Deploy under demand ([#85](https://github.com/prefapp/daggerverse/issues/85)) ([e28b555](https://github.com/prefapp/daggerverse/commit/e28b555dd4da84d0c2335b527284c18c7b480eca))
* fix version on commit ([93153dd](https://github.com/prefapp/daggerverse/commit/93153ddfe255eaa1243e6e094794fc208375a676))


### Bug Fixes

* Clean old files when commiting changes ([#76](https://github.com/prefapp/daggerverse/issues/76)) ([4ee13f2](https://github.com/prefapp/daggerverse/commit/4ee13f2d9288184a3bf1654f7eeece2b92f0eb15))
* commit message ([d71aa27](https://github.com/prefapp/daggerverse/commit/d71aa27c1c06b1d1c09e76615d307cd192295830))
* Panic when an error is found ([#84](https://github.com/prefapp/daggerverse/issues/84)) ([3f95a40](https://github.com/prefapp/daggerverse/commit/3f95a4098da505c35fc814d4ae662ab32d20bf0e))

## [1.1.0](https://github.com/prefapp/daggerverse/compare/hydrate-orchestrator-v1.0.0...hydrate-orchestrator-v1.1.0) (2024-12-20)


### Features

* Add new helmfile render [#49](https://github.com/prefapp/daggerverse/issues/49) ([168f243](https://github.com/prefapp/daggerverse/commit/168f2438435c4d8793c2b270583d14630ea7b3e9))


### Bug Fixes

* Fix render interface ([#68](https://github.com/prefapp/daggerverse/issues/68)) ([0c05cee](https://github.com/prefapp/daggerverse/commit/0c05ceecaf2a3e5ec96bbf1ac41fb3c95acfab1a))
* Ignore environments and repositories files ([#72](https://github.com/prefapp/daggerverse/issues/72)) ([6937821](https://github.com/prefapp/daggerverse/commit/6937821f13fae17de7ab28162e2e3162682328fe))
