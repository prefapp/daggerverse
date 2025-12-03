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
		return "", fmt.Errorf("Error creating the list of imported artifacts: %s", err)
	}

	createdGhResources, err := kindContainer.Directory(
		"/resources/firestartr-crs/github",
	).Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating the list of generated artifacts: %s", err)
	}

	createdInfraResources, err := kindContainer.Directory(
		"/resources/firestartr-crs/infra",
	).Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating the list of generated infra artifacts: %s", err)
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

#### Copied to the cache:
- /import
- /resources
`,
		strings.Join(importedFiles, "\n- "),
		strings.Join(createdGhResources, "\n- "),
		strings.Join(createdInfraResources, "\n- "),
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
		return "", fmt.Errorf("Error creating the list of pushed claims: %s", err)
	}

	pushedClaims := strings.Split(strings.TrimSpace(output), "\n")

	pushedGithubCrs, err := kindContainer.Directory(
		"/resources/firestartr-crs/github",
	).Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating the list of pushed github crs: %s", err)
	}

	pushedInfraCrs, err := kindContainer.Directory(
		"/resources/firestartr-crs/infra",
	).Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating the list of pushed infra crs: %s", err)
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
	`,
		m.Bootstrap.Org,
		strings.Join(pushedClaims, "\n- "),
		m.Bootstrap.Org,
		strings.Join(pushedGithubCrs, "\n- "),
		m.Bootstrap.Org,
		strings.Join(pushedInfraCrs, "\n- "),
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
repo. The PR has been created in the following repo:

%s/pulls


⚠️⚠️⚠️⚠️⚠️ USER INTERVENTION REQUIRED 🛑

You need to review and merge it.

Run a deployment of the new applications machinery by invoking the [action](%s/actions/workflows/generate-deployment-kubernetes.yml), using this coordinates:

%s

And merge the resultant deployment PR

`, prURL, prURL, cardinality,
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
repo. The PRs has been created in the following repos:

- %s/pulls
- %s/pulls

⚠️⚠️⚠️⚠️⚠️ USER INTERVENTION REQUIRED 🛑

You need to review and merge them.

Run a deployment of the new secrets machinery by invoking the [action](%s/actions/workflows/generate-deployment.yml), using this coordinates:

%s

And merge the resultant deployment PR

`, repoArgoCD, repoSysServices, repoSysServices, cardinality,
	)
	return m.UpdateSummaryAndRun(ctx, successMessage)

}
