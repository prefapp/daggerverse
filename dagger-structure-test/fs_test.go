package main

import (
	"context"
	"testing"
)

// Test root directory existence
func TestRootDirectory(t *testing.T) {

	m := DaggerStructureTest{}
	ctx := context.Background()

	ctr := dag.Container().From("alpine")

	filePath := "/"
	shouldExist := true

	res, err := m.AssertFileExistence(ctx, ctr, filePath, shouldExist, AssertFileExistenceOpts{})

	if err != nil || !res {
		t.Fatalf(`root directory should exist`)
	}

}

// Test is directory
func TestIsDir(t *testing.T) {
	
	m := DaggerStructureTest{}
	ctx := context.Background()

	ctr := dag.Container().From("alpine")

	path := "/etc"

	isDir, err := m.IsDir(ctx, ctr, path)

	if err != nil || !isDir {
		t.Fatalf(`/etc should be a directory`)
	}

	path = "/etc/hosts"
	
	isDir, err = m.IsDir(ctx, ctr, path)

	if err != nil || isDir {
		t.Fatalf(`/etc/hosts should not be a directory`)
	}

	path = "/etc/foo"

	_, err = m.IsDir(ctx, ctr, path)

	if err == nil {
		t.Fatalf(`/etc/foo should not exist`)
	}
}
