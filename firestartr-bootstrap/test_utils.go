package main

import (
	"dagger/firestartr-bootstrap/internal/dagger"
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
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			daggerDir = daggerDir.WithNewFile(path, string(content),
				dagger.DirectoryWithNewFileOpts{
					Permissions: 0777,
				})
		}

		return nil
	})

	return daggerDir
}
