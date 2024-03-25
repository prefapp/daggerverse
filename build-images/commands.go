package main

import (
  "context"
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

	// container.Export(ctx, "/tmp/c.tar", ContainerExportOpts{})

	// Publish(
	// 	ctx,
	// 	"ttl.sh/a24b56ef-d667-42a6-b2c9-651637eb1c40",
	// 	ContainerPublishOpts{
	// 		ForcedCompression: Gzip,
	// 		MediaTypes:        Ocimediatypes,
	// 	})

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

