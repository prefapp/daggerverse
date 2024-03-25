package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

func (m *BuildImages) buildFlavour(

	ctx context.Context,

	workDir *Directory,

	flavour string,

	flavourData BuildDataFlavour,

) *Container {

	buildArgs := []BuildArg{}

	for argName, argValue := range flavourData.BuildArgs {

		buildArgs = append(buildArgs, BuildArg{Name: argName, Value: argValue})
	}

	fmt.Printf("%s", buildArgs)

	container := workDir.DockerBuild(DirectoryDockerBuildOpts{
		Dockerfile: "./Dockerfile",
		BuildArgs:  buildArgs,
		Platform:   "linux/amd64",
	})

	return container
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

func getFlavour(buildData *BuildData, flavour string) BuildDataFlavour {

	match, _ := regexp.Match("^(snapshot|release):([a-zA-Z0-9_-]+)", []byte(flavour))

	if !match {

		panic(fmt.Sprintf("Invalid flavour format: %s", flavour))

	}

	splitted := strings.Split(flavour, ":")

	if splitted[0] == "snapshot" {

		return buildData.Snapshots[splitted[1]]

	} else {

		return buildData.Releases[splitted[1]]

	}

}

func getAllFlavours(buildData *BuildData) []string {

	flavours := []string{}

	for flavour := range buildData.Snapshots {

		flavours = append(flavours, "snapshot:"+flavour)

	}

	for flavour := range buildData.Releases {

		flavours = append(flavours, "release:"+flavour)

	}

	return flavours

}
