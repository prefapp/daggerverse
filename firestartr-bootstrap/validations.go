package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

func (m *FirestartrBootstrap) ValidateKindKubernetesConnection(
	ctx context.Context,
	kubeconfig *dagger.Directory,
	kindSvc *dagger.Service,
) error {

	clusterName := "kind"

	ep, err := kindSvc.Endpoint(ctx)
	if err != nil {
		return fmt.Errorf("obtaining the kind-cluster endpoint: %w", err)
	}

	parts := strings.Split(ep, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid endpoint format (expected host:port), got: %q", ep)
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("formatting the kind-cluster port: %w", err)
	}

	_, err = dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "docker", "kubectl", "k9s", "curl", "helm"}).
		WithMountedDirectory("/root/.kube", kubeconfig).
		WithWorkdir("/workspace").
		WithServiceBinding("localhost", kindSvc).
		WithExec([]string{
			"kubectl", "config",
			"set-cluster", fmt.Sprintf("kind-%s", clusterName), fmt.Sprintf("--server=https://localhost:%d", port)},
		).
		WithExec([]string{"kubectl", "cluster-info"}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("connecting to the kind-cluster: %w", err)
	}

	return nil
}

func (m *FirestartrBootstrap) ValidateBootstrapFile(ctx context.Context, bootstrapFile *dagger.File) error {
	schema, err := dag.CurrentModule().Source().File("schemas/bootstrap-file.json").Contents(ctx)
	if err != nil {
		return err
	}

	bootstrapFileContents, err := bootstrapFile.Contents(ctx)
	if err != nil {
		return err
	}

	json, err := yaml.YAMLToJSON([]byte(bootstrapFileContents))
	if err != nil {
		return err
	}

	err = validateDocumentSchema(string(json), schema)
	if err != nil {
		return fmt.Errorf("failed to validate bootstrap file: %w", err)
	}
	return nil
}

func (m *FirestartrBootstrap) ValidateCredentialsFile(ctx context.Context, credentialsFileContents string) error {
	schema, err := dag.CurrentModule().Source().File("schemas/credentials-file.json").Contents(ctx)
	if err != nil {
		return err
	}

	jsonDoc, err := yaml.YAMLToJSON([]byte(credentialsFileContents))
	if err != nil {
		return err
	}

	if err := validateDocumentSchema(string(jsonDoc), schema); err != nil {
		return fmt.Errorf("failed to validate credentials file: %w", err)
	}
	return nil
}

func validateDocumentSchema(document string, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(document)

	res, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		return fmt.Errorf("document is not valid %s", res.Errors())
	}
	return nil
}

func (m *FirestartrBootstrap) ValidatePrefappBotPat(ctx context.Context) error {
	destinationPath := "/tmp/repo"

	gitContainer, err := m.CloneFeaturesRepo(ctx, destinationPath)
	if err != nil {
		return err
	}

	clonedDir := gitContainer.Directory(destinationPath)

	// This final check ensures not only the clone command passed, but the files are accessible.
	_, err = clonedDir.Entries(ctx)
	if err != nil {
		return fmt.Errorf("clone succeeded but failed to read directory contents: %w", err)
	}

	return nil
}

func (m *FirestartrBootstrap) ValidateExistenceOfNeededImages(
	ctx context.Context,
) error {

	slimImage := fmt.Sprintf(
		"ghcr.io/prefapp/gitops-k8s:%s_slim",
		m.Bootstrap.Firestartr.OperatorVersion,
	)

	err := validateExistenceOfImage(ctx, slimImage)
	if err != nil {
		return err
	}

	fullImage := fmt.Sprintf("ghcr.io/prefapp/gitops-k8s:%s", fmt.Sprintf(

		"%s_full-%s",
		m.Bootstrap.Firestartr.OperatorVersion,
		m.Creds.CloudProvider.Name,
	))

	err = validateExistenceOfImage(ctx, fullImage)
	if err != nil {
		return err
	}

	return nil

}

func validateExistenceOfImage(
	ctx context.Context,
	imageRef string,
) error {

	// Use an image that has the 'crane' tool (from Google's container-registry tools)
	craneContainer := dag.Container().
		From("gcr.io/go-containerregistry/crane:latest")

	craneArgs := []string{
		"crane",
		"manifest",
		imageRef,
	}

	_, err := craneContainer.
		WithExec(craneArgs).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("image does not exist: %s", imageRef)
	}

	return nil
}

func (m *FirestartrBootstrap) ValidateCliExistence(
	ctx context.Context,
) error {

	moduleName := fmt.Sprintf("@firestartr/cli@%s", m.Bootstrap.Firestartr.CliVersion)

	npmContainer := dag.Container().
		From("node:20-alpine")

	// 'npm view' queries the metadata. We use '--json' for a faster, less verbose response.
	npmArgs := []string{
		"npm",
		"view",
		moduleName,
		"--json",
	}

	_, err := npmContainer.
		WithExec(npmArgs).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("cli version '%s' does not exist", moduleName)
	}

	return nil

}

func (m *FirestartrBootstrap) ValidateOperatorPat(
	ctx context.Context,
) error {
	owner := fmt.Sprintf("firestartr-%s", m.Bootstrap.Env)
	repo := "app-firestartr"
	tokenSecret := dag.SetSecret(
		"token",
		m.Creds.GithubApp.OperatorPat,
	)

	base := dag.Container().From("alpine/curl").
		WithExec([]string{"apk", "add", "jq"}).
		WithSecretVariable("GITHUB_PAT", tokenSecret)

	// Add jq to extract the 'login' field
	getUserCmd := "https://api.github.com/user | jq -r .login"

	// Execute the command using the secure pattern
	getLogin, err := executeCurlCommand(ctx, base, tokenSecret, getUserCmd)

	username, err := getLogin.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated username: %w", err)
	}

	username = strings.TrimSpace(username)

	if username == "null" || username == "" {
		return fmt.Errorf(
			"authentication failed: the PAT is likely invalid, expired, " +
				"or does not have sufficient read access to the 'user' endpoint",
		)
	}

	// --- Step 2: Check the repository permission for that user ---

	// API Endpoint: GET /repos/:owner/:repo/collaborators/:username/permission
	permissionURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/collaborators/%s/permission",
		owner, repo, username,
	)

	// The command argument for executeCurlCommand is the API call + jq filter.
	getPermissionCmd := fmt.Sprintf("%s | jq -r .permission", permissionURL)

	// Execute curl and pipe the JSON output to jq to extract the 'permission' level
	getPermission, err := executeCurlCommand(
		ctx, base, tokenSecret, getPermissionCmd,
	)
	if err != nil {
		return err
	}

	permission, err := getPermission.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to check repository permissions: %w", err)
	}
	permission = strings.TrimSpace(permission)

	// Valid write access permissions are 'push', 'maintain', or 'admin'.
	switch permission {
	case "push", "maintain", "admin":
		return nil
	case "pull", "triage":
		return fmt.Errorf(
			"User doesn't have write access for repo %s/%s. Permissions are: %s",
			owner, repo, permission,
		)
	default:
		return fmt.Errorf(
			"received an unexpected permission result: '%s'. "+
				"Ensure the repository exists and the user is a "+
				"collaborator (or owner)",
			permission,
		)
	}
}

func (m *FirestartrBootstrap) GithubRepositoryExists(
	ctx context.Context,
	repo string,
	ghToken *dagger.Secret,
) (bool, error) {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return false, err
	}

	ctr, err = ctr.
		WithExec([]string{
			"gh",
			"repo",
			"view",
			fmt.Sprintf("%s/%s", m.Bootstrap.Org, repo),
		}, dagger.ContainerWithExecOpts{
			RedirectStdout: "/tmp/stdout",
			RedirectStderr: "/tmp/stderr",
			Expect:         "ANY",
		}).
		Sync(ctx)
	if err != nil {
		return false, err
	}

	stderr, err := ctr.File("/tmp/stderr").Contents(ctx)
	if err != nil {
		return false, err
	}
	stdout, err := ctr.File("/tmp/stdout").Contents(ctx)
	if err != nil {
		return false, err
	}

	fmt.Printf("stdout: %s\n", stdout)
	fmt.Printf("stderr: %s\n", stderr)

	eC, err := ctr.ExitCode(ctx)
	if err != nil {
		return false, err
	}

	if eC != 0 {

		if strings.Contains(stderr, "Could not resolve to a Repository with the name") {
			return false, nil
		}

		return false, fmt.Errorf("failed to check if repository exists: %s", stderr)
	}

	return true, nil
}

func (m *FirestartrBootstrap) ValidateWebhookDoesntExist(
	ctx context.Context,
	ghToken *dagger.Secret,
	webhookUrl string,
) error {
	ctr, err := m.GhContainer(ctx, ghToken)
	if err != nil {
		return err
	}

	hooksOutput, err := ctr.
		WithExec([]string{
			"gh", "api", fmt.Sprintf("orgs/%s/hooks", m.Bootstrap.Org),
			"--jq", fmt.Sprintf(".[] | select(.config.url == \"%s\") | .id", webhookUrl),
		}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("failed to query GitHub API: %w", err)
	}

	exists := strings.TrimSpace(hooksOutput) != ""
	if exists {
		return fmt.Errorf(
			"a webhook with the URL '%s' already exists in the organization '%s'",
			webhookUrl, m.Bootstrap.Org,
		)
	}

	return nil
}

func (m *FirestartrBootstrap) CheckAlreadyCreatedRepositories(
	ctx context.Context,
	ghToken *dagger.Secret,
) error {
	alreadyCreatedRepos := []string{}

	for idx := range m.Bootstrap.Components {
		component := m.Bootstrap.Components[idx]
		if !component.Skipped {
			repoName := component.Name
			if component.RepoName != "" {
				repoName = component.RepoName
			}
			exists, err := m.GithubRepositoryExists(ctx, repoName, ghToken)
			if err != nil {
				return err
			}
			if exists {
				alreadyCreatedRepos = append(alreadyCreatedRepos, repoName)
			}
		}
	}

	if len(alreadyCreatedRepos) > 0 {
		return fmt.Errorf(
			"the following repositories already exist: %s. Delete them or choose different names to proceed",
			strings.Join(alreadyCreatedRepos, ", "),
		)
	}

	return nil
}
