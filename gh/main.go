package main

import (
	"context"
	"strings"
	"time"

	"gh/internal/dagger"

	"github.com/samber/lo"
)

// Gh is Github CLI module for Dagger
type Gh struct {
	// Configuration for the Github CLI binary
	// +private
	Binary GHBinary

	// Configuration for the Github CLI container
	// +private
	GHContainer GHContainer
}

func New(
	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,

	// Gh plugins
	// +optional
	plugins []GHPlugin,

	// Base container for the Github CLI
	// +optional
	base *dagger.Container,
) *Gh {
	return &Gh{
		Binary: GHBinary{
			Version: version,
		},
		GHContainer: GHContainer{
			Base:    base,
			Token:   token,
			Repo:    repo,
			Plugins: plugins,
		},
	}
}

func (m *Gh) Container(
	ctx context.Context,

	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,

	// Gh plugin names
	// +optional
	pluginNames []string,

	// Gh plugin names
	// +optional
	pluginVersions []string,

) (*dagger.Container, error) {
	file, err := lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).binary(ctx)
	if err != nil {
		return nil, err
	}

	// get the github container configuration
	gc := m.GHContainer

	pluginList := []GHPlugin{}

	for idx, pluginName := range pluginNames {
		pluginVersion := ""

		if idx < len(pluginVersions) {
			pluginVersion = pluginVersions[idx]
		}

		pluginList = append(pluginList, GHPlugin{
			Name:    pluginName,
			Version: pluginVersion,
		})
	}

	// update the container with the given token and repository if provided
	gc = lo.Ternary(token != nil, gc.WithToken(token), gc)
	gc = lo.Ternary(repo != "", gc.WithRepo(repo), gc)
	gc = lo.Ternary(pluginList != nil, gc.WithPlugins(pluginList), gc)

	// get the container object with the given binary
	ctr := gc.container(file)

	return ctr, nil
}

// Run a GitHub CLI command (accepts a single command string without "gh").
func (m *Gh) Run(
	ctx context.Context,

	// Command to run.
	cmd string,

	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,

	// Gh plugin names
	// +optional
	pluginNames []string,

	// Gh plugin names
	// +optional
	pluginVersions []string,

	// disable cache
	// +optional
	// +default=false
	disableCache bool,
) (*dagger.Container, error) {
	ctr, err := m.Container(ctx, version, token, repo, pluginNames, pluginVersions)
	if err != nil {
		return nil, err
	}

	// disable cache if requested
	ctr = lo.Ternary(disableCache, ctr.WithEnvVariable("CACHE_BUSTER", time.Now().String()), ctr)

	// run the command and return the container
	return ctr.WithExec([]string{"sh", "-c", strings.Join([]string{"/usr/local/bin/gh", cmd}, " ")}), nil
}

// Get the GitHub CLI binary.
func (m *Gh) Get(
	ctx context.Context,

	// operating system of the binary
	// +optional
	goos string,

	// architecture of the binary
	// +optional
	goarch string,

	// version of the Github CLI
	// +optional
	version string,
) (*dagger.File, error) {
	return lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).
		WithOS(goos).
		WithArch(goarch).
		binary(ctx)
}

// Create a PR with the current changes using GH
func (m *Gh) CreatePR(
	ctx context.Context,

	// title of the PR
	title string,

	// body text of the PR
	body string,

	// branch name
	branch string,

	// path to the repo
	repoDir *dagger.Directory,

	// version of the Github CLI
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,
) (*dagger.Container, error) {
	contentsDirPath := "/content"
	ctr, err := m.Container(ctx, version, token, "", []string{}, []string{})
	if err != nil {
		panic(err)
	}

	ctr = ctr.
		WithMountedDirectory(contentsDirPath, repoDir).
		WithWorkdir(contentsDirPath).
		WithExec([]string{
			"gh",
			"pr",
			"create",
			"--title", title,
			"--body", body,
			"--head", branch,
		})

	_, err = ctr.Sync(ctx)

	if err != nil {
		panic(err)
	}

	return ctr, nil
}

// Commit current changes into a new/existing branch
func (m *Gh) Commit(
	ctx context.Context,

	// path to the repo
	repoDir *dagger.Directory,

	// name of the branch to commit to
	branchName string,

	// commit message
	commitMessage string,

	// GitHub token
	token *dagger.Secret,

	// create a commit even if there are no changes
	// +optional
	// +default=false
	allowEmpty bool,

	// version of the Github CLI
	// +optional
	version string,
) (*dagger.Container, error) {
	contentsDirPath := "/content"
	ctr, err := m.Container(ctx, version, token, "", []string{"prefapp/gh-commit"}, []string{"v1.2.3"})
	if err != nil {
		panic(err)
	}

	ctr = ctr.
		WithMountedDirectory(contentsDirPath, repoDir).
		WithWorkdir(contentsDirPath).
		WithExec([]string{
			"gh",
			"commit",
			"-b", branchName,
			"-m", commitMessage,
			// "--delete-path", fmt.Sprintf("tfworkspaces/%s/%s/%s", tfDep.ClaimName, tfDep.Tenant, tfDep.Environment),
		})

	_, err = ctr.Sync(ctx)

	if err != nil {
		panic(err)
	}

	return ctr, nil
}

// Commit current changes into a new/existing branch
func (m *Gh) CommitAndCreatePR(
	ctx context.Context,

	// path to the repo
	repoDir *dagger.Directory,

	// name of the branch to commit to
	branchName string,

	// commit message
	commitMessage string,

	// title of the PR
	prTitle string,

	// body text of the PR
	prBody string,

	// version of the Github CLI
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,
) (*dagger.Container, error) {
	_, err := m.Commit(ctx, repoDir, branchName, commitMessage, token, false, version)
	if err != nil {
		panic(err)
	}

	return m.CreatePR(ctx, prTitle, prBody, branchName, repoDir, version, token)
}
