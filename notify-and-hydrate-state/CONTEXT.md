# notify-and-hydrate-state

Analyzes claim pull request changes and triggers hydration of the wet repository.

## Language

**Claim PR**:
A pull request in the claims repository that adds, modifies, or deletes one or
more Claims.
_Avoid_: PR, claims pull request

**Orphan PR**:
A pull request in the wet repository whose originating claim PR has already been
closed or merged, leaving it without a live parent.
_Avoid_: stale PR, dangling PR

**Diff result**:
The categorized set of files changed by a claim PR: added, modified, deleted,
and unmodified.
_Avoid_: changeset, delta
