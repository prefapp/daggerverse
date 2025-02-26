# Changelog

## [2.0.0](https://github.com/prefapp/daggerverse/compare/hydrate-kubernetes-v1.0.0...hydrate-kubernetes-v2.0.0) (2025-02-26)


### ⚠ BREAKING CHANGES

* Removed app parameter from hydrate-orchestrator
* Merge pull request #113 from prefapp/fix/remove-app-parameter-from-orchestrator

### Features

* [state-repo] Deploy under demand ([#85](https://github.com/prefapp/daggerverse/issues/85)) ([e28b555](https://github.com/prefapp/daggerverse/commit/e28b555dd4da84d0c2335b527284c18c7b480eca))
* Add sys service namespace ([#101](https://github.com/prefapp/daggerverse/issues/101)) ([db44a4e](https://github.com/prefapp/daggerverse/commit/db44a4ef6956ceddd68cf27e53866ecc00237911))
* jsonpatch for image keys ([#131](https://github.com/prefapp/daggerverse/issues/131)) ([7ade374](https://github.com/prefapp/daggerverse/commit/7ade3749b1cc6aa7a818086d2dc80918daccfb03))
* Merge pull request [#113](https://github.com/prefapp/daggerverse/issues/113) from prefapp/fix/remove-app-parameter-from-orchestrator ([5cf8816](https://github.com/prefapp/daggerverse/commit/5cf8816b651c5cd7e345cb0ab29640ce7fdc041d))
* support firestartr configs, generate  `respositories.yaml` and  `environments.yaml` in runtime ([#96](https://github.com/prefapp/daggerverse/issues/96)) ([79d9cd9](https://github.com/prefapp/daggerverse/commit/79d9cd96cb37637f23751a87aa3c06802f1ad94b))
* support for release name and images file priority ([#88](https://github.com/prefapp/daggerverse/issues/88)) ([ec49028](https://github.com/prefapp/daggerverse/commit/ec4902885cfc88e61a933918c7149cf18bd1b59b))
* support images file on values dir ([#82](https://github.com/prefapp/daggerverse/issues/82)) ([4a627ed](https://github.com/prefapp/daggerverse/commit/4a627edfe0eda86f3818a701fbbc8d7452611071))
* support multiple helm credentials with `~/.config/helm/registry/config.json` ([#90](https://github.com/prefapp/daggerverse/issues/90)) ([07733a2](https://github.com/prefapp/daggerverse/commit/07733a2db842a5e79b1c6680db691d22dcde28d2))
* support sys services ([#92](https://github.com/prefapp/daggerverse/issues/92)) ([858c21d](https://github.com/prefapp/daggerverse/commit/858c21d7114ecca78fd0a017daa5df2ed6fe3992))


### Bug Fixes

* images matrix ([#115](https://github.com/prefapp/daggerverse/issues/115)) ([1f15ed1](https://github.com/prefapp/daggerverse/commit/1f15ed1108bfe8e84dfd2363517088af733bf109))
* Inferred app name from state repo param ([ad246a9](https://github.com/prefapp/daggerverse/commit/ad246a9b78c4a0ad24c30d82fb7ec86fd17c35f0))
* Removed app parameter from hydrate-orchestrator ([5cf8816](https://github.com/prefapp/daggerverse/commit/5cf8816b651c5cd7e345cb0ab29640ce7fdc041d))
* tpl sets ([#98](https://github.com/prefapp/daggerverse/issues/98)) ([bb328dd](https://github.com/prefapp/daggerverse/commit/bb328dd193df0f8c70e907488ec05c4aee615e23))

## 1.0.0 (2024-12-20)


### Features

* Add new helmfile render [#49](https://github.com/prefapp/daggerverse/issues/49) ([168f243](https://github.com/prefapp/daggerverse/commit/168f2438435c4d8793c2b270583d14630ea7b3e9))
* enforce unit testing ([#51](https://github.com/prefapp/daggerverse/issues/51)) ([fb04c89](https://github.com/prefapp/daggerverse/commit/fb04c891e788a32c71e5c7355b2b32a06a30a02b))
* hydrate-kubernetes module ([#47](https://github.com/prefapp/daggerverse/issues/47)) ([2b89969](https://github.com/prefapp/daggerverse/commit/2b89969f0b589639cce3d76c626b6fdafa906cce))
* render extra artifacts with chart artifacts ([#73](https://github.com/prefapp/daggerverse/issues/73)) ([37e5d80](https://github.com/prefapp/daggerverse/commit/37e5d802e46c109eabb7a9087439a834b4930bd9))
* sys apps rendering and kubernetes render config ([#65](https://github.com/prefapp/daggerverse/issues/65)) ([07b0b9f](https://github.com/prefapp/daggerverse/commit/07b0b9f0ffaf3400aa5665bf2dd2bc00d7110402))
* upgrade dagger engine version v0.15.1 ([#52](https://github.com/prefapp/daggerverse/issues/52)) ([2d8b4de](https://github.com/prefapp/daggerverse/commit/2d8b4de5d77f1207cea7f0aed663a2fc4b6a014a))


### Bug Fixes

* change interface to adapt to the hydrate-orchestrator ([#60](https://github.com/prefapp/daggerverse/issues/60)) ([d9df238](https://github.com/prefapp/daggerverse/commit/d9df2386b0d9bf5ee32adeebdeb48166cb707cbb))
* execs check go template ([#50](https://github.com/prefapp/daggerverse/issues/50)) ([b44498c](https://github.com/prefapp/daggerverse/commit/b44498c261ac5b61fa53e2118d5f4e8252becd63))
* ignore empty resources ([#71](https://github.com/prefapp/daggerverse/issues/71)) ([eebb795](https://github.com/prefapp/daggerverse/commit/eebb7959af38c200817b3d547c2df12820205cea))
* password as stdin in helm login ([4422b44](https://github.com/prefapp/daggerverse/commit/4422b44ed482a4b65469e5c6f1e56383fc6c2789))
* remove unused attr ([#74](https://github.com/prefapp/daggerverse/issues/74)) ([cecdcbc](https://github.com/prefapp/daggerverse/commit/cecdcbcc3dc4a50cc37979b3f6e4c5a9d11a7131))
* rename sys-apps to sys-services ([#69](https://github.com/prefapp/daggerverse/issues/69)) ([763e2fd](https://github.com/prefapp/daggerverse/commit/763e2fd9d1319f9d7243c42ff070dc1abe4c5548))
