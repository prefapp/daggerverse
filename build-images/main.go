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
	"log"
	"strings"
)

func (m *BuildImages) BuildImageBatch(

	ctx context.Context,

	yamlPath *File,

	workDir *Directory,

	// +optional
	// +default="all"
	flavours string,

	// +optional
	publish bool,

	// +optional
	address string,

) string {

	builds := []*Container{}

	flavoursList := []string{}

	buildData := loadInfo(ctx, yamlPath)

	if flavours == "all" {

		flavoursList = getAllFlavours(buildData)

	} else {

		flavoursList = strings.Split(flavours, ",")
	}

	for _, flavour := range flavoursList {

		builds = append(builds, m.BuildFlavour(ctx, workDir, yamlPath, flavour))

		log.Println("Built flavour: ", flavour)
	}

	if publish {

		for _, image := range builds {

			val, err := image.Publish(ctx, address)

			if err != nil {

				log.Panic(fmt.Sprintf("Error publishing image: %s", val))
			}
		}
	}

	return strings.Join(flavoursList, ",")

}

func (m *BuildImages) BuildFlavour(

	ctx context.Context,

	workDir *Directory,

	yamlPath *File,

	flavour string,

) *Container {

	buildData := loadInfo(ctx, yamlPath)

	flavourData := getFlavour(buildData, flavour)

	return m.buildFlavour(
		ctx,

		workDir,

		flavour,

		flavourData)

}
