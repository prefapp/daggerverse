package main

import (
	"context"
	"gh/internal/dagger"

	"github.com/samber/lo"
)

type GHPlugin struct {
	Name    string
	Version string
}

type GHContainer struct {
	// Base container for the Github CLI
	Base *dagger.Container

	// Github token
	Token *dagger.Secret

	// Github Repository
	Repo string

	// Github Plugins
	Plugins []GHPlugin
}

// WithRepo returns the GHContainer with the given repository.
func (c GHContainer) WithRepo(repo string) GHContainer {
	return GHContainer{
		Base:    c.Base,
		Token:   c.Token,
		Repo:    repo,
		Plugins: c.Plugins,
	}
}

// WithToken returns the GHContainer with the given token.
func (c GHContainer) WithToken(token *dagger.Secret) GHContainer {
	return GHContainer{
		Base:    c.Base,
		Token:   token,
		Repo:    c.Repo,
		Plugins: c.Plugins,
	}
}

// WithPlugin returns the GHContainer with the given plugin.
func (c GHContainer) WithPlugins(plugins []GHPlugin) GHContainer {
	return GHContainer{
		Base:    c.Base,
		Token:   c.Token,
		Repo:    c.Repo,
		Plugins: plugins,
	}
}

// container returns the container for the Github CLI with the given binary.
func (c GHContainer) container(binary *dagger.File) *dagger.Container {
	return lo.Ternary(c.Base != nil, c.Base, dag.Container().From("alpine/git:latest")).
		WithFile("/usr/local/bin/gh", binary).
		// WithEntrypoint([]string{"/usr/local/bin/gh"}).
		WithEnvVariable("GH_PROMPT_DISABLED", "true").
		WithEnvVariable("GH_NO_UPDATE_NOTIFIER", "true").
		With(func(ctr *dagger.Container) *dagger.Container {
			if c.Token != nil {
				token, err := c.Token.Plaintext(context.Background())
				if err != nil {
					panic(err)
				}

				ctr = ctr.WithExec([]string{"gh", "auth", "login", "--with-token"}, dagger.ContainerWithExecOpts{
					Stdin: token,
				}).WithExec([]string{"gh", "auth", "setup-git"})
			}

			if c.Repo != "" {
				ctr = ctr.WithEnvVariable("GH_REPO", c.Repo)
			}

			if c.Plugins != nil {
				// for each plugin, add the plugin to the container
				for _, pluginData := range c.Plugins {
					if pluginData.Name == "" {
						panic("Unnamed plugin found")
					}

					command := []string{
						"gh", "extension", "install", pluginData.Name,
					}

					if pluginData.Version != "" {
						command = append(command, "--pin", pluginData.Version)
					}

					ctr = ctr.WithExec(command)
				}
			}

			return ctr
		})
}
