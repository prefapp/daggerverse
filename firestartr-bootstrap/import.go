package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
)

func (m *FirestartrBootstrap) RunImporter(
	ctx context.Context,
	dockerSocket *dagger.Socket,
	kindSvc *dagger.Service,
	claimsDir *dagger.Directory,
	crsDir *dagger.Directory,
) *dagger.Container {

	claimsDefaults, err := m.RenderClaimsDefaults(ctx,
		dag.CurrentModule().
			Source().
			File("firestartr_files/claims/.config/claims_defaults.tmpl"),
	)
	if err != nil {
		return nil
	}

	defaultsDir := dag.Directory().
		WithNewDirectory("/claims_defaults").
		WithNewFile("claims_defaults.yaml", claimsDefaults)

	kindContainer := GetKind(dockerSocket, kindSvc).
		WithDirectory("/claims", claimsDir).
		WithDirectory("/crs", crsDir).
		WithDirectory("/config", dag.CurrentModule().Source().Directory("firestartr_files/crs/.config")).
		WithDirectory("/claims_defaults", defaultsDir).
		WithExec([]string{"apk", "add", "nodejs", "npm"}).
		WithExec([]string{"npm", "install", "-g", fmt.Sprintf("@firestartr/cli@%s", m.Bootstrap.Firestartr.Version)}).
		WithExec(
			[]string{
				"firestartr-cli", "importer",
				"--claims", "/claims",
				"--claimsDefaults", "/claims_defaults",
				"--config", "/config",
				"--org", m.GhOrg,
				"--crs", "/crs",
				"--filters", "gh-members,REGEXP=[A-Za-z0-9\\-]+",
				"--filters", "gh-group,REGEXP=[A-Za-z0-9\\-]+",
				"--filters", "gh-repo,SKIP=SKIP",
			},
		)

	kindContainer = m.ApplyFirestartrCrs(ctx, kindContainer)

	return kindContainer

}
