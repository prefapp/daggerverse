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

func (m *HydrateKubernetes) BuildHelmRepositoriesFile(

	ctx context.Context,

	dotFirestartrDir *dagger.Directory,

	envFileLocation string,

) (*dagger.File, error) {

	envFile := m.ValuesDir.File(envFileLocation)

	fsConf := dag.FirestartrConfig(dotFirestartrDir)

	regs, err := fsConf.Registries(ctx)

	if err != nil {

		return nil, err

	}

	envYamlContent, err := envFile.Contents(ctx)

	if err != nil {

		return nil, err

	}

	envYamlStruct := EnvYaml{}

	err = yaml.Unmarshal([]byte(envYamlContent), &envYamlStruct)

	if err != nil {

		return nil, err

	}

	var helmRepos []HelmRepo

	repositoryName := strings.Split(envYamlStruct.Chart, "/")[0]

	var hRepo HelmRepo

	if envYamlStruct.Registry != "" {

		hRepo = HelmRepo{

			Name: repositoryName,

			Url: envYamlStruct.Registry,
		}

		helmRepos = append(helmRepos, hRepo)

	} else {

		for _, reg := range regs {

			regName, err := reg.Name(ctx)

			if err != nil {

				return nil, err

			}

			regHost, err := reg.Registry(ctx)

			if err != nil {

				return nil, err

			}

			if repositoryName == regName {

				hRepo = HelmRepo{

					Name: regName,

					Url: regHost,

					Oci: true,
				}

				helmRepos = append(helmRepos, hRepo)

			}

		}

		if len(helmRepos) == 0 {

			return nil, fmt.Errorf("No registry found for repository in your firestartr config directory %s", repositoryName)
		}
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
