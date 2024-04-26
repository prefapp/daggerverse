package main

import (
	"context"
	"strconv"
	"strings"
)

type AssertFileExistenceOpts struct {
	Permissions *string
	Uid         *int
	Gid         *int
}

func (m *DaggerStructureTest) AssertFileExistence(ctx context.Context, container *Container, path string, shouldExist bool, options AssertFileExistenceOpts) (bool, error) {

	isDir, err := m.IsDir(ctx, container, path)

	if err != nil && shouldExist {
		return false, err
	}

	var ctr *Container

	if isDir {
		ctr, err = dag.Container().From("alpine").WithDirectory("/mnt", container.Directory(path)).Sync(ctx)
	} else {
		ctr, err = dag.Container().From("alpine").WithFile("/mnt", container.File(path)).Sync(ctx)
	}

	if err != nil {
		return false, err
	}

	stat, err := ctr.WithExec([]string{"stat", "-c", "%A %U %G", "/mnt"}).Stdout(ctx)

	s := strings.Split(stat, " ")

	permissions, uid, gid := s[0], s[1], s[2]

	if err != nil {
		return false, err
	}

	if options.Permissions != nil && permissions != *options.Permissions {
		return false, nil
	}

	if options.Uid != nil && uid != strconv.Itoa(*options.Uid) {
		return false, nil
	}

	if options.Gid != nil && gid != strconv.Itoa(*options.Gid) {
		return false, nil
	}
	
	return true, nil
}

func (m *DaggerStructureTest) IsDir(ctx context.Context, container *Container, path string) (bool, error) {

	_, err := container.Directory(path).Sync(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not a directory") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
