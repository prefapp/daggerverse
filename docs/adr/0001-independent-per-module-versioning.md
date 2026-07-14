# Independent per-module versioning via release-please

Each Dagger module in this monorepo has its own independent version, managed via
`release-please` with separate package entries in `release-please-config.json`.
This was chosen because modules have different release cadences and are consumed
independently (e.g. `github.com/prefapp/daggerverse/kind@v0.3.1`) — a single
monorepo version would force every consumer to update on every unrelated module
change and would mask which module actually changed. The alternative (one version
for the whole repo) was rejected because it couples independent module consumers
unnecessarily.
