package main

import (
	"context"
	"dagger/hydrate-kubernetes/internal/dagger"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

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

	hydrateK8sConf, err := valuesDir.
		File(".github/hydrate_k8s_config.yaml").
		Contents(ctx)

	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal([]byte(hydrateK8sConf), config)

	if err != nil {

		return nil, err

	}

	c := dag.
		Container().
		From(config.Image)

	c = containerWithCmds(c, config.Commands)

	if helmfile == nil {

		helmfile = dag.CurrentModule().Source().File("./helm-" + renderType + "/helmfile.yaml")

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
