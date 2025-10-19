package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Check for new images here: https://github.com/helmfile/helmfile/pkgs/container/helmfile
var HELMFILE_DOCKER_IMAGE = "ghcr.io/helmfile/helmfile:v1.1.0"

type HydrateKubernetes struct {
	Container               *dagger.Container
	ValuesDir               *dagger.Directory
	WetRepoDir              *dagger.Directory
	Helmfile                *dagger.File
	ValuesGoTmpl            *dagger.File
	HelmRegistryLoginNeeded bool
	HelmConfigDir           *dagger.Directory
	RenderType              string
	DotFirestartrDir        *dagger.Directory
	RepositoriesFile        *dagger.File
}

func New(

	ctx context.Context,

	// The path to the values directory, where the helm values are stored
	valuesDir *dagger.Directory,

	// The path to the wet repo directory, where the wet manifests are stored
	wetRepoDir *dagger.Directory,

	// The path to the helmfile.yaml file
	// +optional
	helmfile *dagger.File,

	// The path to the values.go.tmpl file
	// +optional
	valuesGoTmpl *dagger.File,

	// +optional
	// +default=false
	helmRegistryLoginNeeded bool,

	// +optional
	helmConfigDir *dagger.Directory,

	// Type of the render, it can be apps or sys-apps
	// +optional
	// +default="apps"
	renderType string,

	// Firestartr org directory, it should lives in the
	dotFirestartr *dagger.Directory,

) (*HydrateKubernetes, error) {

	image := HELMFILE_DOCKER_IMAGE

	extraCommands := [][]string{}

	confFileExists, err := hydrateConfigFileExists(ctx, valuesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check for hydrate_k8s_config.yaml: %w", err)
	}

	config := &Config{}
	if confFileExists {
		hydrateK8sConf, err := valuesDir.
			File(".github/hydrate_k8s_config.yaml").
			Contents(ctx)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal([]byte(hydrateK8sConf), config)
		if err != nil {
			return nil, err
		}

		if config.Image != "" {
			image = config.Image
		}

		if config.Commands != nil {
			extraCommands = config.Commands
		}
	}

	c := dag.Container().From(image)

	c = containerWithCmds(c, extraCommands)

	if helmfile == nil {
		helmfile = dag.CurrentModule().
			Source().
			File("./helm-" + renderType + "/helmfile.yaml.gotmpl")
	}

	if valuesGoTmpl == nil {
		valuesGoTmpl = dag.CurrentModule().Source().File("./helm-" + renderType + "/values.yaml.gotmpl")
	}

	return &HydrateKubernetes{

		Container: c,

		ValuesDir: valuesDir,

		WetRepoDir: wetRepoDir,

		Helmfile: helmfile,

		DotFirestartrDir: dotFirestartr,

		ValuesGoTmpl: valuesGoTmpl,

		HelmConfigDir: helmConfigDir,

		RenderType: strings.Trim(
			strings.ToLower(renderType),
			" ",
		),
	}, nil
}

// This function renders the apps or sys-apps based on the render type
// It returns the wet directory where the rendered files are stored
func (m *HydrateKubernetes) Render(

	ctx context.Context,

	app string,

	cluster string,

	// +optional
	// +default=""
	tenant string,

	//+optional
	//+default=""
	env string,

	// +optional
	// +default="{\"images\":[]}"
	newImagesMatrix string,

) ([]*dagger.Directory, error) {

	if m.RenderType == "sys-services" {

		dir, err := m.DumpSysAppRenderToWetDir(ctx, app, cluster)

		return []*dagger.Directory{dir}, err

	} else if m.RenderType == "apps" {

		if tenant == "" || env == "" {

			panic("--tenant and --env are required params for apps render")

		}

		dir, err := m.DumpAppRenderToWetDir(
			ctx,
			app,
			cluster,
			tenant,
			env,
			newImagesMatrix,
		)

		return []*dagger.Directory{dir}, err

	} else {

		panic(
			fmt.Sprintf(
				"Invalid render type: %s, it should be either 'apps' or 'sys-services'",
				m.RenderType,
			),
		)
	}
}
