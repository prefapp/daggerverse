# Update claims' features

A Dagger module intended to run in GitHub Actions that updates feature claims to the latest release.

This module either updates feature version fields to the latest available release (from the `features` repository) and opens a PR in the `claims` repository, or triggers the hydration workflow when a feature uses a `ref` field.

## Key behaviors

- If a feature in a claim has a `version` field, the module checks it against the newest release available in the `features` repository, updates the field if needed, and creates a PR in the `claims` repository only when an update is required.
- If a feature has a `ref` field, the module will not change the claim; instead it will trigger the hydration workflow for that claim.
- If a claim contains multiple features and some use `version` while others use `ref`, the module creates a PR updating only the `version` features and does not trigger hydration automatically (hydration must run after the PR is merged).
- If a PR is created, the workflow can optionally automerge it when the `automerge` option is enabled.

## Validation

Before making changes, the claim(s) are validated against the official JSON Schema to avoid creating invalid changes. The module uses the claims schema hosted in the firestartr-pro/docs repository:

https://github.com/firestartr-pro/docs/blob/main/site/raw/core/claims/claims.schema.json

If validation fails for a claim, the workflow reports the error in the workflow summary and will not create a PR or trigger the hydration workflow for that claim.

## Automerge behavior

When automerge is requested the workflow requests an automated merge (`gh pr merge --auto`). The PR will be merged automatically once required checks pass and merge conditions are satisfied.
