package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
)

func (m *FirestartrBootstrap) ExtractClaimFromCache(
	ctx context.Context,
	cacheVolume *dagger.CacheVolume,
	claimName string,
) (*dagger.Directory, error) {
	script := fmt.Sprintf(`
import os, sys, yaml, shutil

claim_name = %q

yaml_count = 0
yml_count = 0
claim_matches = 0
cr_matches = 0
claim_files = []

for root, dirs, files in os.walk("/cache"):
    for f in files:
        path = os.path.join(root, f)
        try:
            with open(path) as fh:
                data = yaml.safe_load(fh)
        except:
            continue
        if not isinstance(data, dict):
            continue

        rel = os.path.relpath(path, "/cache")

        if f.endswith(".yaml"):
            yaml_count += 1
        elif f.endswith(".yml"):
            yml_count += 1

        if (f.endswith(".yaml") or f.endswith(".yml")) and "/claims/" in rel.replace("\\", "/"):
            if data.get("name") == claim_name:
                dest = os.path.join("/output", os.path.basename(path))
                os.makedirs("/output", exist_ok=True)
                shutil.copy2(path, dest)
                claim_matches += 1
                claim_files.append(os.path.basename(path))

        if f.endswith(".yaml") or f.endswith(".yml"):
            annotations = data.get("metadata", {}).get("annotations", {})
            if isinstance(annotations, dict):
                ref = annotations.get("firestartr.dev/claim-ref")
                if ref is not None:
                    ref_name = ref.split("/")[-1]
                    if ref_name == claim_name:
                        dest = os.path.join("/output", os.path.basename(path))
                        os.makedirs("/output", exist_ok=True)
                        shutil.copy2(path, dest)
                        cr_matches += 1
                    else:
                        print("  cr with ref '%%s' != '%%s': %%s" %% (ref, claim_name, rel))
            elif "/firestartr-crs/" in rel.replace("\\", "/"):
                print("  cr file with no annotations or non-dict: %%s" %% rel)

print("Scanned %%d .yaml files, %%d .yml files" %% (yaml_count, yml_count))
print("Found %%d claim matches, %%d CR matches" %% (claim_matches, cr_matches))
`, claimName)

	ctr, err := dag.Container().
		From("python:3-alpine").
		WithMountedCache("/cache", cacheVolume).
		WithExec([]string{"pip", "install", "--quiet", "pyyaml"}).
		WithExec([]string{"mkdir", "-p", "/output"}).
		WithNewFile("/search.py", script).
		WithExec([]string{"python", "/search.py"}).
		Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search cache: %s", extractErrorMessage(err, "unknown error during cache search"))
	}

	claimEntries, err := ctr.Directory("/output").Glob(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list output: %w", err)
	}

	if len(claimEntries) == 0 {
		return nil, fmt.Errorf(
			"no claim file found with name %q in cache. "+
				"Ensure the cache volume was populated by running the import step first",
			claimName,
		)
	}

	return ctr.Directory("/output"), nil
}
