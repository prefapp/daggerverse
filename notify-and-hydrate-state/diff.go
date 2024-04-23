package main

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"
)

type DiffResult struct {
	AddedFiles []*File

	DeletedFiles []*File

	ModifiedFiles []*File
}

func (m *NotifyAndHydrateState) CompareDirs(

	ctx context.Context,

	local Directory,

	remote Directory,

) DiffResult {

	localEntries, err := local.Entries(ctx)

	if err != nil {

		panic(err)

	}

	remoteEntries, err := remote.Entries(ctx)

	if err != nil {

		panic(err)

	}

	result := DiffResult{}

	entriesInBothDirs := []string{}

	for _, localEntry := range localEntries {

		remoteIndex := slices.Index(remoteEntries, localEntry)

		if remoteIndex != -1 {

			entriesInBothDirs = append(entriesInBothDirs, localEntry)

			// These two lines delete the element from the slice
			remoteEntries[remoteIndex] = remoteEntries[len(remoteEntries)-1]

			remoteEntries = remoteEntries[:len(remoteEntries)-1]

		} else {

			result.AddedFiles = append(result.AddedFiles, local.File(localEntry))

		}

	}

	for _, remoteEntry := range remoteEntries {

		result.DeletedFiles = append(result.DeletedFiles, remote.File(remoteEntry))

	}

	for _, entry := range entriesInBothDirs {

		localContents, err := local.File(entry).Contents(ctx)

		if err != nil {

			panic(err)

		}

		remoteContents, err := remote.File(entry).Contents(ctx)

		if err != nil {

			panic(err)

		}

		if localContents != remoteContents {

			result.ModifiedFiles = append(result.ModifiedFiles, local.File(entry))

		}

	}

	fmt.Println(" ----- ADDED ----- ")

	PrintFileList(ctx, result.AddedFiles)

	fmt.Println(" ----- DELETED ----- ")

	PrintFileList(ctx, result.DeletedFiles)

	fmt.Println(" ----- MODIFIED ----- ")

	PrintFileList(ctx, result.ModifiedFiles)

	return result
}

func PrintFileList(

	ctx context.Context,

	listToPrint []*File,

) {

	for _, file := range listToPrint {

		contents, err := file.Contents(ctx)

		if err != nil {

			panic(err)

		}

		fmt.Print(contents)

	}

}
