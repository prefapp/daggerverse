# firestartr-config

Models the `.firestartr` configuration directory that drives Firestartr deployments and validates it against the Firestartr-owned dot-firestartr contracts.

## Schema validation

Configuration under `apps/`, `platforms/`, `docker_registries/`, and `helm_repositories/` is validated **before** decoding/using it against the corresponding contract schemas:

- `apps.schema.json`
- `platforms.schema.json`
- `docker_registries.schema.json`
- `helm_repositories.schema.json`

Runtime schemas are sourced from the published npm package
`@firestartr/firestartr-claims_schemas` (`dot-firestartr/<contract>.schema.json`),
which is the authoritative upstream distribution for these contracts.

### Source selection, versioning and pinning policy

- `schemaVersion` (default: `v3.2.0`) selects the schema release. Pass an explicit npm semver version (e.g. `v3.2.0` or `3.2.0`, normalized for npm) to pin deterministically, or pass a dist-tag such as `latest` to deliberately opt into the latest published schemas (non-deterministic).
- `schemaDirectory` (optional, `*dagger.Directory`): **portable offline/pre-fetched schema source**. When provided, schemas are read via the Dagger API from `<schemaVersion>/<contract>.schema.json` inside the directory and it takes precedence over `schemaBaseURL`. This is the supported way to use schemas available on the caller's machine (pass them in as a directory), since the module runtime cannot read caller-host filesystem paths.
- `schemaBaseURL` (optional) selects the remote schema source when `schemaDirectory` is not provided:
  - empty or `"npm"` (default): resolve schemas from the npm registry tarball for `@firestartr/firestartr-claims_schemas`.
  - an `http(s)` URL: treated as an npm registry base (for mirrors/testing), resolving the same package from that registry.
- The module pins `v3.2.0` by default. If the pinned npm package/version is not available from the selected registry, resolution fails fast with an actionable error; provide `schemaDirectory` with pre-fetched schemas to validate offline.

### Resolution, caching and failure behavior

- npm source: package metadata and the version tarball are resolved over HTTP(S), tarball integrity is verified (`dist.integrity`, with `dist.shasum` fallback), and `package/dot-firestartr/*.schema.json` contracts are extracted in-memory (pure Go, no Node/npm runtime dependency).
- directory source: contract schemas are read from `schemaDirectory` via the Dagger API at `<schemaVersion>/<contract>.schema.json`.
- Each provider load resolves/fetches the npm tarball (or reads directory schemas) and compiles contract schemas once per `New` invocation and caches them in-memory for that load (no per-file/per-contract refetch).
- Fetch/read/resolve/parse/compile/verification failures and validation failures are surfaced as actionable diagnostics including the file (when applicable), contract, requested/resolved version and schema source (package/registry/metadata/tarball or schema directory and path), and cause/reason. Invalid configuration (including registries) fails loading; invalid entries are never silently dropped.
- Runtime validation uses the selected `schemaVersion`/source, then decodes into the generated pinned contract types. If a selected schema (e.g. `latest` or a different version) diverges incompatibly from the pinned/generated shape, validation/decoding may fail with an actionable diagnostic rather than silently dropping required data.

## Schema-derived types

Contract model types (`FirestartrApp`, `FirestartrService`, `Registry`, `BasePaths`, `Repository`, `Platform`, `AllowedClaim`, `AuthStrategy`) are **generated from the pinned `v3.2.0` schema snapshot** and must not be hand-edited.

- Source snapshot: `testdata/dot-firestartr/v3.2.0/*.schema.json`
- Generator: `tools/gentypes` (invoked via `generate.go` / `go:generate`)
- Generated output: `types_gen.go`

To refresh derived types from the pinned schemas:

```sh
cd firestartr-config
go generate ./...
dagger develop
```

When bumping the default pinned schema version, update the snapshot in `testdata/dot-firestartr/<version>/`, update `DefaultSchemaVersion`/generator inputs, regenerate types, and run tests. Runtime validation always uses the explicitly selected `schemaVersion`/`schemaBaseURL`; decoding targets the generated (pinned) types after successful validation. If a selected schema (e.g. `latest`) diverges incompatibly from the pinned/generated shape, validation/decoding may fail with an actionable diagnostic rather than silently dropping required data.

## Language

**Platform**:
A named deployment target (e.g. a Kubernetes cluster) with a type, a set of
tenants and environments, and constraints on which Claim resource types and
providers are allowed.
_Avoid_: cluster, environment, target

**App**:
A Firestartr-managed application grouping services for a given tenant and
environment.
_Avoid_: workload, service, deployment

**Registry**:
A container image registry configuration used by Firestartr-managed deployments.
_Avoid_: image repository, image source

**dotFirestartr directory**:
The `.firestartr/` directory containing the platform, app, registry, and
repository YAML configurations consumed by this module.
_Avoid_: config dir, values dir

## Compatibility impact

- Constructor is schema-aware: `New(dotFirestartr, schemaVersion?, schemaBaseURL?, schemaDirectory?)` with default pinned `schemaVersion` `v3.2.0`.
- Runtime schema source defaults to the npm package `@firestartr/firestartr-claims_schemas` resolved via registry tarball (pure Go, integrity-verified). `schemaBaseURL` selects the remote source (empty/`"npm"` = npm registry, `http(s)` = npm registry/mirror base). For offline/pre-fetched schemas, pass `schemaDirectory` (`*dagger.Directory`) containing `<schemaVersion>/<contract>.schema.json` (read via the Dagger API); it takes precedence over `schemaBaseURL`. Caller-host filesystem paths in `schemaBaseURL` are not a supported offline fallback (the module runtime cannot read them). Legacy docs raw URL file-base loading is not supported.
- Apps: service definitions use `services` (no longer `image_types`); valid service lists deserialize as `[{repo, service_names, platforms?}]`.
- Docker registries: required `name`, `registry` (host only, no scheme/port), `image_types`, `default`, `auth_strategy` (`aws_oidc|azure_oidc|generic|ghcr|dockerhub|none`), and `base_paths` with required `services` (optional `charts`). Invalid registries now fail validation instead of being silently dropped.
- Helm repositories: require non-empty `name` and `url` beginning with `http://`, `https://`, or `oci://`.
- Platforms: `platforms/` remains canonical and both `.yaml` and `.yml` files are supported/validated.
- Downstream Dagger consumers must regenerate bindings/migrate to the updated API and config shapes/semantics (separate consumer migration). If a selected runtime schema (e.g. `latest`) diverges incompatibly from the pinned/generated types, validation/decoding may fail with an actionable diagnostic.
