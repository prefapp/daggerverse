# Update claims' features

This dagger module updates a feature's claim to the latest version. This can be done in one of two ways:

- If the feature has a `version` field, then that field's value will be updated to the latest avaliable version, which we get by pooling the releases of the `features` repository on GitHub, and a PR with the updated claim will be created in the `claims` repository.
- If the feature has a `ref` field, no changes will be made. Instead, the hydration workflow will be automatically called.
- If the claim has multiple features to be updated and some have `version` fields while others have `ref` fields, then just the PR (updating only the features with `version` fields) will be created, and the hydration workflow will not be automatically called. This is because the hydration workflow must be called after merging the PR anyway.

Whether a PR was created or a workflow was called, the link to it will be posted in the workflow summary.

The workflow also validates claims by using the JSON Schema located at the [`firestartr-pro/docs` repository](https://github.com/firestartr-pro/docs/blob/main/site/raw/core/claims/claims.schema.json), to avoid errors caused by missing fields or wrong field types. If a claim is invalid, it will be reported in the workflow summary, and no PR will be created or workflow called for that claim.
