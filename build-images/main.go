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
)

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

) *Container {

	containers := []*Container{}

	buildData := loadInfo(ctx, yamlPath)

	// return m.buildFlavour(ctx, workDir, "default", buildData.Releases["default"])

	// for flavour, flavourData := range buildData["snapshots"] {

	// 	m.BuildFlavour(workDir, flavour, flavourData)
	// }

	for flavour, flavourData := range buildData.Releases {

		containers = append(

			containers,

			m.buildFlavour(ctx, workDir, flavour, flavourData),
		)

	}

	return containers[0]

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

    workDir,

    flavour,

    flavourData

  )

}


