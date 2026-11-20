package main

import (
	"dagger/firestartr-config/internal/dagger"
	"io/fs"
	"os"
	"path/filepath"
)

// This is a helper function that reads the files from a directory and returns a dagger.Directory
// Dagger cannot read files from testing go system with dag.CurrentModule(),
// so we need to read the files and create a simulated dagger.Directory
func getDir(dirPath string) *dagger.Directory {
	daggerDir := dag.Directory()

	filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		rel, err := filepath.Rel(dirPath, path)
		if err != nil {
			panic(err)
		}

		daggerDir = daggerDir.WithNewFile(rel, string(content), dagger.DirectoryWithNewFileOpts{
			Permissions: 0o777,
		})

		return nil
	})

	return daggerDir
}

func testFixturesDir() *dagger.Directory {
	return getDir("./fixtures/.firestartr")
}

func testSchemaProvider() *schemaProvider {
	p, err := newSchemaProvider(DefaultSchemaVersion, "./testdata/dot-firestartr")
	if err != nil {
		panic(err)
	}
	return p
}
