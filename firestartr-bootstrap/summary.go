package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type ExecutionState struct {
	CompletedSteps int
	LastStepName   string
	LogID          string
	Status         string
}

var updatedMarkdownContent string

func (m *FirestartrBootstrap) UpdateSummaryAndRun(
	ctx context.Context,
	stepDescription string,
) string {

	var currentMarkdown string

	if updatedMarkdownContent == "" {

		currentMarkdown = fmt.Sprintf("# 🚀 BootstrapFile Execution Summary\n\n---\n")
	} else {

		currentMarkdown = updatedMarkdownContent
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	newMarkdownEntry := fmt.Sprintf(
		"\n### ✅ Step Completed: %s\n"+
			"* **Time:** %s\n"+
			"* **Status:** Success\n"+
			"---\n",
		stepDescription,
		timestamp,
	)

	updatedMarkdownContent = currentMarkdown + newMarkdownEntry

	return updatedMarkdownContent
}

func (m *FirestartrBootstrap) UpdateSummaryAndRunForImportResourcesStep(
	ctx context.Context,
	kindContainer *dagger.Container,
) (string, error) {

	importedFiles, err := kindContainer.Directory("/import/crs").Entries(
		ctx,
	)
	if err != nil {
		return "", fmt.Errorf("error creating the list of imported artifacts: %w", err)
	}

	createdGhResources, err := kindContainer.Directory(
		"/resources/firestartr-crs/github",
	).Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("error creating the list of generated artifacts: %w", err)
	}

	createdInfraResources, err := kindContainer.Directory(
		"/resources/firestartr-crs/infra",
	).Glob(ctx, "FirestartrTerraformWorkspace.*")
	if err != nil {
		return "", fmt.Errorf("error creating the list of generated infra artifacts: %w", err)
	}

	createdSecretResources, err := kindContainer.Directory(
		"/resources/firestartr-crs/infra",
	).Glob(ctx, "ExternalSecret.*")
	if err != nil {
		return "", fmt.Errorf("error creating the list of generated secret artifacts: %w", err)
	}

	successMessage := fmt.Sprintf(`
=====================================================
📥 RESOURCES IMPORTED AND CREATED 📥
=====================================================
Initial CRs checked, missing resources created, and
all necessary configurations copied to the cache volume.
The environment is fully provisioned.

#### Imported resources:
- %s

#### Generated and created resources (GitHub):
- %s

#### Generated and created resources (Infra):
- %s

#### Generated and created resources (Secrets):
- %s

#### Copied to the cache:
- /import
- /resources
`,
		strings.Join(importedFiles, "\n- "),
		strings.Join(createdGhResources, "\n- "),
		strings.Join(createdInfraResources, "\n- "),
		strings.Join(createdSecretResources, "\n- "),
	)

	return m.UpdateSummaryAndRun(ctx, successMessage), nil

}

func (m *FirestartrBootstrap) UpdateSummaryAndRunForPushResourcesStep(
	ctx context.Context,
	kindContainer *dagger.Container,
) (string, error) {

	cmd := []string{"find", "/resources/claims", "-type", "f", "-name", "*.yaml"}

	output, err := kindContainer.
		WithExec(cmd).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("error creating the list of pushed claims: %w", err)
	}

	pushedClaims := strings.Split(strings.TrimSpace(output), "\n")

	pushedGithubCrs, err := kindContainer.Directory(
		"/resources/firestartr-crs/github",
	).Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("error creating the list of pushed GitHub CRs: %w", err)
	}

	pushedInfraCrs, err := kindContainer.Directory(
		"/resources/firestartr-crs/infra",
	).Glob(ctx, "FirestartrTerraformWorkspace.*")
	if err != nil {
		return "", fmt.Errorf("error creating the list of pushed infra CRs: %w", err)
	}

	pushedSecretCrs, err := kindContainer.Directory(
		"/resources/firestartr-crs/infra",
	).Glob(ctx, "ExternalSecret.*")
	if err != nil {
		return "", fmt.Errorf("error creating the list of pushed secret CRs: %w", err)
	}

	successMessage := fmt.Sprintf(`
	=====================================================
				⤴️RESOURCE PUSH COMPLETE ⤴️
	=====================================================
	GitHub access machinery initialized, resources copied
	from cache, and all configuration files pushed to
	the destination environment.

#### List of pushed Claims (%s/claims)
	- %s

#### List of pushed CRs (%s/state-github)
	- %s

#### List of pushed CRs (%s/state-infra)
	- %s

#### List of pushed CRs (%s/state-secrets)
	- %s
	`,
		m.Bootstrap.Org,
		strings.Join(pushedClaims, "\n- "),
		m.Bootstrap.Org,
		strings.Join(pushedGithubCrs, "\n- "),
		m.Bootstrap.Org,
		strings.Join(pushedInfraCrs, "\n- "),
		m.Bootstrap.Org,
		strings.Join(pushedSecretCrs, "\n- "),
	)
	return m.UpdateSummaryAndRun(ctx, successMessage), nil

}

func (m *FirestartrBootstrap) UpdateSummaryAndRunForRollbackStep(
	ctx context.Context,
	deletionSummary string,
) string {

	successMessage := fmt.Sprintf(`
=====================================================
⏪ ROLLBACK OPERATION COMPLETE ⏪
=====================================================
Artifacts processed successfully. The system state
has been reverted to the previous stable configuration.

#### List of deletions
%s
`,
		deletionSummary,
	)

	return m.UpdateSummaryAndRun(ctx, successMessage)

}

func (m *FirestartrBootstrap) ShowSummaryReport(
	ctx context.Context,
) string {

	fmt.Print(updatedMarkdownContent)

	return updatedMarkdownContent
}

func (m *FirestartrBootstrap) UpdateSummaryAndRunForPushDeploymentStep(
	ctx context.Context,
	prURL string,
	cardinality string,
) string {

	successMessage := fmt.Sprintf(`
=====================================================
     🚀 DEPLOYMENT FIRESTARTR PUSHED 🚀
=====================================================
Deployment files created and successfully pushed to the app-firestartr
repo. The PR has been created in its corresponding repo.


⚠️⚠️⚠️⚠️⚠️ USER INTERVENTION REQUIRED 🛑


You need to perform the following actions:

  1. Review and merge the PR created in:
     - %s/pulls

  2. Run a deployment of the new applications machinery by invoking the [action](%s/actions/workflows/generate-deployment-kubernetes.yml), using these coordinates:

     %s

  3. Review and merge the resultant deployment PR that is created by the action.

`,
		prURL,
		prURL,
		cardinality,
	)
	return m.UpdateSummaryAndRun(ctx, successMessage)

}

func (m *FirestartrBootstrap) UpdateSummaryAndRunForPushArgoCDStep(
	ctx context.Context,
	repoArgoCD string,
	repoSysServices string,
	cardinality string,
) string {

	successMessage := fmt.Sprintf(`
=====================================================
     🚀 DEPLOYMENT ARGOCD APPLICATIONS PUSHED 🚀
=====================================================
Application and secret files created and successfully pushed to the argocd
repo. The PRs have been created in their corresponding repos.


⚠️⚠️⚠️⚠️⚠️ USER INTERVENTION REQUIRED 🛑


You need to perform the following actions:

  1. Review and merge the PRs created in:
     - %s/pulls
     - %s/pulls

  2. Run a deployment of the new secrets machinery by invoking the [action](%s/actions/workflows/generate-deployment.yml), using these coordinates:

     %s

  3. Review and merge the resultant deployment PR that is created by the action.

`,
		repoArgoCD,
		repoSysServices,
		repoSysServices,
		cardinality,
	)
	return m.UpdateSummaryAndRun(ctx, successMessage)

}
