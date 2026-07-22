# Cross-module dependency strategy: local paths for co-developed modules, pinned remote refs otherwise

Modules that are co-developed in this repo and released together use local
relative paths in `dagger.json` (e.g. `../kind`, `../firestartr-config`).
Modules consuming a stable, already-published version use pinned remote refs
(e.g. `github.com/prefapp/daggerverse/gh@<sha>`). Local paths allow synchronized
development without publishing intermediate versions; pinned remote refs give
consumers reproducible builds. The mixing is intentional and not an
inconsistency: local for tight co-development, pinned remote for stable
consumption.
