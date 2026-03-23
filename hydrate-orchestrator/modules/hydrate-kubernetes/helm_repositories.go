package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type HelmRepo struct {
	Name     string `yaml:"name"`
	Url      string `yaml:"url"`
	Username string `yaml:"username,omitempty"`
	Oci      bool   `yaml:"oci,omitempty"`
}

type RepositoriesStructFile struct {
	Repositories []HelmRepo `yaml:"repositories"`
}

func (m *HydrateKubernetes) getDeploymentConfig(
	ctx context.Context,

	envFileLocation string,
) (*EnvYaml, error) {

	envFile := m.ValuesDir.File(envFileLocation)

	envYamlContent, err := envFile.Contents(ctx)

	if err != nil {

		return nil, err

	}

	envYamlStruct := EnvYaml{}

	err = yaml.Unmarshal([]byte(envYamlContent), &envYamlStruct)

	if err != nil {

		return nil, err

	}

	return &envYamlStruct, nil
}

/*
getHelmReposFromFirestartrConfig collects all Helm repository configurations
from the Firestartr config directory and returns them as a slice of HelmRepo.
It aggregates all possible Helm repositories that can be used by the Helm
charts in the Firestartr config directory into a single repositories
representation.
*/
func (m *HydrateKubernetes) getHelmReposFromFirestartrConfig(
	ctx context.Context,

	dotFirestartrDir *dagger.Directory,

	deploymentConfig *EnvYaml,
) ([]HelmRepo, error) {

	// Resulting array of helm repositories
	var helmRepos []HelmRepo

	repositoryName := strings.Split(deploymentConfig.Chart, "/")[0]

	fsConf := dag.FirestartrConfig(dotFirestartrDir)

	legacyRegs, err := fsConf.LegacyRegistries(ctx)

	if err != nil {
		return nil, err
	}

	regs, err := fsConf.Registries(ctx)

	if err != nil {

		return nil, err

	}

	for _, reg := range regs {

		regName, err := reg.Name(ctx)

		if err != nil {

			return nil, err

		}

		regHost, err := reg.URL(ctx)

		if err != nil {

			return nil, err

		}

		if repositoryName == regName {

			hRepo := HelmRepo{

				Name: regName,

				Url: regHost,

				Oci: true,
			}

			helmRepos = append(helmRepos, hRepo)

		}
	}

	repositories, err := fsConf.Repositories(ctx)

	if err != nil {

		return nil, err

	}

	for _, repo := range repositories {

		repoName, err := repo.Name(ctx)

		if err != nil {

			return nil, err

		}

		repoUrl, err := repo.URL(ctx)

		if err != nil {

			return nil, err

		}

		if repositoryName == repoName {

			helmRepos = append(helmRepos, HelmRepo{
				Name: repoName,
				Url:  repoUrl,
				Oci:  false,
			})
		}
	}

	for _, legacyReg := range legacyRegs {

		name, err := legacyReg.Name(ctx)
		if err != nil {
			return nil, err
		}

		url, err := legacyReg.Registry(ctx)
		if err != nil {
			return nil, err
		}

		if repositoryName == name {
			helmRepos = append(helmRepos, HelmRepo{
				Name: name,
				Url:  url,
				Oci:  true,
			})
		}
	}

	if len(helmRepos) == 0 {

		return nil, fmt.Errorf("No registry found for repository in your firestartr config directory %s", repositoryName)
	}

	return helmRepos, nil
}

func (m *HydrateKubernetes) BuildHelmRepositoriesFile(

	ctx context.Context,

	dotFirestartrDir *dagger.Directory,

	envFileLocation string,

) (*dagger.File, error) {

	envConfig, err := m.getDeploymentConfig(ctx, envFileLocation)

	if err != nil {
		return nil, err
	}

	helmRepos, err := m.getHelmReposFromFirestartrConfig(
		ctx,
		dotFirestartrDir,
		envConfig,
	)
	if err != nil {

		return nil, err

	}

	reposStruct := RepositoriesStructFile{

		Repositories: helmRepos,
	}

	reposStructYamlContent, err := yaml.Marshal(reposStruct)

	if err != nil {

		return nil, err

	}

	return dag.Directory().
		WithNewFile("repositories.yaml", string(reposStructYamlContent)).
		File("repositories.yaml"), nil
}
