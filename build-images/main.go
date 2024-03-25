// A generated module for Common functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

type BuildImages struct{}

type BuildData struct {
	Snapshots map[string]BuildDataFlavour `yaml:"snapshots,flow"`

	Releases map[string]BuildDataFlavour `yaml:"releases,flow"`
}

type BuildDataFlavour struct {
	BuildArgs map[string]string `yaml:"build_args,flow"`

	Dockerfile string

	Tag string
}

func (m *BuildImages) BuildImageBatch(

	ctx context.Context,

	yamlPath *File,

	workDir *Directory,

	// +optional
	publish bool,

	// +optional
	address string,

) {

}

func (m *BuildImages) BuildImage(

	ctx context.Context,

	yamlPath *File,

	// +optional
	// +default="releases,default"
	flavour string,

	workDir *Directory,

	publish bool,

	address string,

) []string {

	containers := []string{}

	buildData := loadInfo(ctx, yamlPath)

	// return m.buildFlavour(ctx, workDir, "default", buildData.Releases["default"])

	// for flavour, flavourData := range buildData["snapshots"] {

	// 	m.BuildFlavour(workDir, flavour, flavourData)
	// }

	for flavour, flavourData := range buildData.Releases {

		flav, err := m.buildFlavour(ctx, workDir, flavour, flavourData)

		if err != nil {

			panic(fmt.Sprintf("Building flavour: %s", flavour))

		}

		containers = append(

			containers,

			flav,
		)

	}

	return containers

}

func (m *BuildImages) buildFlavour(

	ctx context.Context,

	workDir *Directory,

	flavour string,

	flavourData BuildDataFlavour,

) (string, error) {

	buildArgs := []BuildArg{}

	for argName, argValue := range flavourData.BuildArgs {

		buildArgs = append(buildArgs, BuildArg{Name: argName, Value: argValue})
	}

	fmt.Printf("%s", buildArgs)

	flavour, error := workDir.
		DockerBuild(DirectoryDockerBuildOpts{
			Dockerfile: flavourData.Dockerfile,
			BuildArgs:  buildArgs,
			Platform:   "linux/amd64",
		}).
		Publish(
			ctx,
			"https://ttl.sh/",
			ContainerPublishOpts{
				ForcedCompression: Gzip,
				MediaTypes:        Ocimediatypes,
			})

	if error != nil {

		return "", error

	}

	return flavour, nil
	// return dag.Container().BuildDocker(workDir, ContainerBuildOpts{

	// 	Dockerfile: "./Dockerfile",

	// 	Target: "",

	// 	BuildArgs: buildArgs,
	// })
}

func loadInfo(ctx context.Context, yamlPath *File) *BuildData {

	val, err := yamlPath.Contents(ctx)

	buildData := BuildData{}

	if err != nil {

		panic(fmt.Sprintf("Loading yaml: %s", val))

	} else {

		err := yaml.Unmarshal([]byte(val), &buildData)

		if err != nil {

			panic(fmt.Sprintf("cannot unmarshal data: %v", err))

		}

		return &buildData

	}

}
