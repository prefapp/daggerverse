package main

import (
	"context"
	"fmt"
	"regexp"
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

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,

) (*dagger.Container, error) {
	file, err := lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).binary(ctx, localGhCliPath, token)
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

	if version != "" && localGhCliPath != nil {
		versionOutput, err := ctr.WithExec([]string{"gh", "--version"}).Stdout(ctx)
		if err != nil {
			return nil, err
		}

		// versionOutput[0] example => "gh version <version> (<date>)"
		// this regexp removes the " (<date>)" portion of the output
		dateText := regexp.MustCompile(`\s\([0-9-]+\)`)
		currentVersion := dateText.ReplaceAllString(
			strings.ReplaceAll(
				strings.Split(versionOutput, "\n")[0],
				"gh version ",
				"",
			),
			"",
		)

		if currentVersion != version {
			fmt.Printf(
				"WARNING: local gh binary version and specified version differ. Local gh version: %s, specified version: %s",
				currentVersion, version,
			)
		}
	}

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

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,
) (*dagger.Container, error) {
	ctr, err := m.Container(ctx, version, token, repo, pluginNames, pluginVersions, localGhCliPath)
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

	// GitHub token.
	// +optional
	token *dagger.Secret,
) (*dagger.File, error) {
	return lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).
		WithOS(goos).
		WithArch(goarch).
		binary(ctx, nil, token)
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

	// name of the branch used as the base for the commit
	// +optional
	baseBranch string,

	// labels to add to the PR
	// +optional
	labels []string,

	// reviewers to add to the PR
	// +optional
	reviewers []string,

	// version of the Github CLI
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,

	// container with gh binary
	// +optional
	ctr *dagger.Container,

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,
) (string, error) {
	contentsDirPath := "/content"

	var err error

	if ctr == nil {
		ctr, err = m.Container(ctx, version, token, "", []string{}, []string{}, localGhCliPath)
		if err != nil {
			panic(err)
		}
	}

	cmd := []string{
		"gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--head", branch,
	}

	if baseBranch != "" {
		cmd = append(cmd, "--base", baseBranch)
	}

	for _, label := range labels {
		cmd = append(cmd, "--label", label)
	}

	filteredReviewers := m.filterReviewers(reviewers)
	for _, reviewer := range filteredReviewers {
		cmd = append(cmd, "--reviewer", reviewer)
	}

	ctr = ctr.
		WithMountedDirectory(contentsDirPath, repoDir).
		WithWorkdir(contentsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec(cmd)

	_, err = ctr.Sync(ctx)

	if err != nil {
		panic(err)
	}

	prId, err := ctr.
		WithExec([]string{
			"gh", "pr", "list",
			"--head", branch,
			"--json", "number",
			"--jq", ".[0].number",
		}).
		Stdout(ctx)

	if err != nil {
		panic(err)
	}

	prLink, err := ctr.
		WithExec([]string{
			"gh", "pr", "view",
			"--json", "url",
			"--jq", ".url",
			strings.TrimSpace(prId),
		}).
		Stdout(ctx)

	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(prLink), nil
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

	// name of the branch used as the base for the commit
	// +optional
	baseBranch string,

	// delete-path parameter for gh commit plugin
	// +optional
	deletePath string,

	// create an empty commit
	// +optional
	// +default=false
	createEmpty bool,

	// version of the Github CLI
	// +optional
	version string,

	// container with gh binary
	// +optional
	ctr *dagger.Container,

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,
) (*dagger.Container, error) {
	contentsDirPath := "/content"

	var err error

	if ctr == nil {
		ctr, err = m.Container(
			ctx,
			version,
			token,
			"",
			[]string{"prefapp/gh-commit"},
			[]string{"v1.3.0"},
			localGhCliPath,
		)
		if err != nil {
			panic(err)
		}
	}

	ctr = ctr.
		WithMountedDirectory(contentsDirPath, repoDir).
		WithWorkdir(contentsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String())

	cmd := []string{
		"gh", "commit",
		"-b", branchName,
		"-m", commitMessage,
		"--delete-path", deletePath,
	}

	if baseBranch != "" {
		cmd = append(cmd, "--base", baseBranch)
	}

	if createEmpty {
		cmd = append(cmd, "-e")
	}

	ctr = ctr.WithExec(cmd)

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

	// name of the branch used as the base for the commit
	// +optional
	baseBranch string,

	// labels to add to the PR
	// +optional
	labels []string,

	// reviewers to add to the PR
	// +optional
	reviewers []string,

	// delete-path parameter for gh commit plugin
	// +optional
	deletePath string,

	// create an empty commit
	// +optional
	// +default=false
	createEmpty bool,

	// version of the Github CLI
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,
) (string, error) {
	ctr, err := m.Container(
		ctx,
		version,
		token,
		"",
		[]string{"prefapp/gh-commit"},
		[]string{"v1.3.0"},
		localGhCliPath,
	)
	if err != nil {
		panic(err)
	}

	m.DeleteRemoteBranch(ctx, repoDir, branchName, version, token, ctr, localGhCliPath)

	ctr, err = m.Commit(
		ctx, repoDir, branchName, commitMessage, token, baseBranch,
		deletePath, createEmpty, version, ctr, localGhCliPath,
	)
	if err != nil {
		panic(err)
	}

	return m.CreatePR(
		ctx, prTitle, prBody, branchName, repoDir, baseBranch,
		labels, reviewers, version, token, ctr, localGhCliPath,
	)
}

func (m *Gh) DeleteRemoteBranch(
	ctx context.Context,

	// path to the repo
	repoDir *dagger.Directory,

	// name of the branch to commit to
	branchName string,

	// version of the Github CLI
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *dagger.Secret,

	// container with gh binary
	// +optional
	// default=nil
	ctr *dagger.Container,

	// runner's gh dir path
	// +optional
	localGhCliPath *dagger.File,
) {
	contentsDirPath := "/content"

	var err error

	if ctr == nil {
		ctr, err = m.Container(
			ctx,
			version,
			token,
			"",
			[]string{},
			[]string{},
			localGhCliPath,
		)
		if err != nil {
			panic(err)
		}
	}

	ctr = ctr.
		WithMountedDirectory(contentsDirPath, repoDir).
		WithWorkdir(contentsDirPath).
		WithEnvVariable("CACHE_BUSTER", time.Now().String())

	remoteBranchList, err := ctr.
		WithExec([]string{"git", "ls-remote"}).
		Stdout(ctx)
	if err != nil {
		panic(err)
	}

	exp := regexp.MustCompile(fmt.Sprintf("%s\n", branchName))
	matches := exp.Match([]byte(remoteBranchList))

	if matches {
		_, err = ctr.
			WithExec([]string{"git", "push", "-d", "origin", branchName}).
			Sync(ctx)

		if err != nil {
			panic(err)
		}
	}
}
