package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"strings"

	"github.com/samber/lo"
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

// Create a function that get's all helm possible configurations from the firestartr config directory and creates a helm repositories file with all the possible helm repositories that can be used in the helm charts of the firestartr config directory and return a helmrepo structure array.
 func (m *HydrateKubernetes) getHelmReposFromFirestartrConfig(
	ctx context.Context,

	dotFirestartrDir *dagger.Directory,

	deploymentConfig *EnvYaml,
 ) ([]HelmRepo, error) {

	// Resulting array of helm repositories
	var helmRepos []HelmRepo

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

		regHost, err := reg.Registry(ctx)

		if err != nil {

			return nil, err

		}

		repositoryName := strings.Split(deploymentConfig.Chart, "/")[0]

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

		helmRepos = append(helmRepos, HelmRepo{
			Name: repoName,
			Url: repoUrl,
			Oci: false,
		})
	}

	/* Keep legacy logic*/
	repositoryName := strings.Split(deploymentConfig.Chart, "/")[0]

	if deploymentConfig.Registry != "" {
		helmRepos = append(helmRepos, HelmRepo{
			Name: repositoryName,
			Url: deploymentConfig.Registry,
			Oci: lo.FromPtrOr(deploymentConfig.Oci, false),
		})
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

		hr := HelmRepo{
			Name: name,
			Url: url,
			Oci: true,
		}

		helmRepos = append(helmRepos, hr)
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


	envConfig , err := m.getDeploymentConfig(ctx, envFileLocation)
	
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

	// panic(string(reposStructYamlContent))

	return dag.Directory().
		WithNewFile("repositories.yaml", string(reposStructYamlContent)).
		File("repositories.yaml"), nil
}
