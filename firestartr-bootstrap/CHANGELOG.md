# Changelog

## [1.3.0](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.2.5...firestartr-bootstrap-v1.3.0) (2026-02-27)


### Features

* add boot example configurations ([#408](https://github.com/prefapp/daggerverse/issues/408)) ([076c543](https://github.com/prefapp/daggerverse/commit/076c543d8184340800a86915b92c8856b2fd3bca))
* Allow `latest` for operator and cli versions ([#412](https://github.com/prefapp/daggerverse/issues/412)) ([9b5ecbc](https://github.com/prefapp/daggerverse/commit/9b5ecbc3984b1613155b511a650302b246900551))


### Bug Fixes

* [`firestartr-bootstrap`] Use controller GitHub App instead of Admin one ([#411](https://github.com/prefapp/daggerverse/issues/411)) ([b302839](https://github.com/prefapp/daggerverse/commit/b302839764613b006b46bc17ecde2b0bdfb01c30))
* better handling errors ([#387](https://github.com/prefapp/daggerverse/issues/387)) ([ffba81e](https://github.com/prefapp/daggerverse/commit/ffba81e36b0686e27f9bfe054fa9fdd1c552edee))
* Error when patching &lt;org&gt;-all claim ([#398](https://github.com/prefapp/daggerverse/issues/398)) ([ff3274f](https://github.com/prefapp/daggerverse/commit/ff3274f90122b0bae4cb87794ae10c2cf790b66b))
* Update docs with version field changes ([#396](https://github.com/prefapp/daggerverse/issues/396)) ([d2d837d](https://github.com/prefapp/daggerverse/commit/d2d837d07646bf13175e3ebe1c84511bc390f142))
* Upload missing rego validation policy ([#401](https://github.com/prefapp/daggerverse/issues/401)) ([030d6db](https://github.com/prefapp/daggerverse/commit/030d6db4284f84c990ee3c3fc52afcb7367e7fd7))
* Wait on each resource kind in a group rather than individually ([#409](https://github.com/prefapp/daggerverse/issues/409)) ([160f532](https://github.com/prefapp/daggerverse/commit/160f5328cfb994a01701548c38ffe2935d36b208))
* Wrong app label and duplicated folder ([#421](https://github.com/prefapp/daggerverse/issues/421)) ([a323bd6](https://github.com/prefapp/daggerverse/commit/a323bd6d62211530107f6ac68016f9c073cb6592))

## [1.2.5](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.2.4...firestartr-bootstrap-v1.2.5) (2026-01-13)


### Bug Fixes

* Better handling of dagger.WithExec() errors ([#379](https://github.com/prefapp/daggerverse/issues/379)) ([f933d7c](https://github.com/prefapp/daggerverse/commit/f933d7c38a5447c4ec1420ae72473fa279caa12d))
* Update documentation to include new script flags ([#384](https://github.com/prefapp/daggerverse/issues/384)) ([cf7d802](https://github.com/prefapp/daggerverse/commit/cf7d802b5406053ce3332c446909cf2d9f9c1039))
* Update step_by_step.sh to not reuse clusters ([#382](https://github.com/prefapp/daggerverse/issues/382)) ([a22cf85](https://github.com/prefapp/daggerverse/commit/a22cf854cd32fefd31e42feb7373e4790d992bb6))
* Wrong label on step_by_step.sh ([#383](https://github.com/prefapp/daggerverse/issues/383)) ([f2dfae5](https://github.com/prefapp/daggerverse/commit/f2dfae59d135baf2c239f804925f9d05a3c03563))

## [1.2.4](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.2.3...firestartr-bootstrap-v1.2.4) (2025-12-31)


### Bug Fixes

* Allow downloading latest version of a feature by not setting a version or setting 'latest' ([#375](https://github.com/prefapp/daggerverse/issues/375)) ([4b81367](https://github.com/prefapp/daggerverse/commit/4b81367e692a53bb8ed3034114f90938de6563dd))
* Check webhook doesn't exist before attempting the creation of resources ([#376](https://github.com/prefapp/daggerverse/issues/376)) ([de36ef8](https://github.com/prefapp/daggerverse/commit/de36ef854a5d0e3649a090d40faf65dd80e97063))
* Delete repositories on rollback ([#372](https://github.com/prefapp/daggerverse/issues/372)) ([0f5f049](https://github.com/prefapp/daggerverse/commit/0f5f049caf4763ae7c894a332781b13c4b2009fd))
* Enable PR creation and approving globally for org ([#378](https://github.com/prefapp/daggerverse/issues/378)) ([e5b65db](https://github.com/prefapp/daggerverse/commit/e5b65db7f66d7ce9d29ce1c1b786d0fa0580a6d9))

## [1.2.3](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.2.2...firestartr-bootstrap-v1.2.3) (2025-12-15)


### Bug Fixes

* **firestartr-bootstrap:** updated CRDs URL in local-operator bash script ([#367](https://github.com/prefapp/daggerverse/issues/367)) ([0708df0](https://github.com/prefapp/daggerverse/commit/0708df094aca6b23c9c10b732e8e299cd36f2156))

## [1.2.2](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.2.1...firestartr-bootstrap-v1.2.2) (2025-12-11)


### Bug Fixes

* Updated ArgoCD secrets to also be remoteRefs to secrets ([#364](https://github.com/prefapp/daggerverse/issues/364)) ([eb95e24](https://github.com/prefapp/daggerverse/commit/eb95e24a5a17c5322707b65fc21b30727357e5f9))

## [1.2.1](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.2.0...firestartr-bootstrap-v1.2.1) (2025-12-09)


### Bug Fixes

* Missing actions.oidc section ([#359](https://github.com/prefapp/daggerverse/issues/359)) ([02c9bdc](https://github.com/prefapp/daggerverse/commit/02c9bdc25a400a5609ead6a405f369bfbc9860d4))

## [1.2.0](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.1.0...firestartr-bootstrap-v1.2.0) (2025-12-05)


### Features

* add validations ([#351](https://github.com/prefapp/daggerverse/issues/351)) ([712e29e](https://github.com/prefapp/daggerverse/commit/712e29e9dc199cc19f9f8295e1cb81ea6f4d22b4))
* Add webhook creation support ([#339](https://github.com/prefapp/daggerverse/issues/339)) ([a48e7bc](https://github.com/prefapp/daggerverse/commit/a48e7bc854c1690b848b2cf6df9b036b6bd7838d))
* Added webhookUrl config parameter ([#346](https://github.com/prefapp/daggerverse/issues/346)) ([615802a](https://github.com/prefapp/daggerverse/commit/615802a133028343c47e7c9b68607b4cee7dd559))
* bootstrap add new deployment ([#355](https://github.com/prefapp/daggerverse/issues/355)) ([f608466](https://github.com/prefapp/daggerverse/commit/f608466ca5e6bc16acdb6a573a78829b3e24f75a))

## [1.1.0](https://github.com/prefapp/daggerverse/compare/firestartr-bootstrap-v1.0.0...firestartr-bootstrap-v1.1.0) (2025-10-09)


### Features

* Adapt bootstrap to 1.48 ([#316](https://github.com/prefapp/daggerverse/issues/316)) ([2effa28](https://github.com/prefapp/daggerverse/commit/2effa285fc5d914cbf71e58b40768b43ef750b4c))
* Get credentials from parameter store ([#331](https://github.com/prefapp/daggerverse/issues/331)) ([d6e647d](https://github.com/prefapp/daggerverse/commit/d6e647d84c3f856db24c7bdc239ef8658b152bb5))


### Bug Fixes

* Implement general improvements ([#319](https://github.com/prefapp/daggerverse/issues/319)) ([ee87ae7](https://github.com/prefapp/daggerverse/commit/ee87ae7d59dca7650aa1e3465ca7ee89698c431b))
* More documentation fixes ([#330](https://github.com/prefapp/daggerverse/issues/330)) ([1150685](https://github.com/prefapp/daggerverse/commit/11506857e4b11b75bedee3dedd7cbddff243cfc2))
* Small docs fix ([dec12f5](https://github.com/prefapp/daggerverse/commit/dec12f56f1a0c648ad99ced7cc3f26af66adfdc9))
* Update README docs ([#329](https://github.com/prefapp/daggerverse/issues/329)) ([df69f46](https://github.com/prefapp/daggerverse/commit/df69f46e410f6e2a569c10874c307f63d8233803))

## 1.0.0 (2025-07-02)


### Features

* firestartr bootstrap docs ([#205](https://github.com/prefapp/daggerverse/issues/205)) ([a88314f](https://github.com/prefapp/daggerverse/commit/a88314f02683d7bd9b4c77cb822062f9398cad3f))
* firestartr bootstrap validations ([#193](https://github.com/prefapp/daggerverse/issues/193)) ([d961fa6](https://github.com/prefapp/daggerverse/commit/d961fa6f651641c5ce5b52059047e91fadd83019))
* firestartr-bootstrap ([#185](https://github.com/prefapp/daggerverse/issues/185)) ([94a6a09](https://github.com/prefapp/daggerverse/commit/94a6a096e25347e539164290887b4088d8ec2250))
