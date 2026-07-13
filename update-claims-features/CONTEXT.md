# update-claims-features

Keeps feature references in Claims up to date by checking for new releases and
opening PRs or triggering hydration.

## Language

**Version pin**:
A `version` field in a claim's feature entry referencing a specific release tag
from the features repository. The module checks this against the latest release
and updates it when outdated.
_Avoid_: version field, version lock

**Ref pin**:
A `ref` field in a claim's feature entry referencing a branch or commit SHA
instead of a release tag. The module never modifies a ref-pinned feature; it
triggers hydration directly instead.
_Avoid_: ref field, branch ref

**Version constraint**:
An optional semver constraint (e.g. `~1.2`) that limits which releases the
module will upgrade a version pin to.
_Avoid_: version filter, semver range
