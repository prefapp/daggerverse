package main

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"
)

func (m *NotifyAndHydrateState) CompareDirs(

	ctx context.Context,

	local Directory,

	remote Directory,

) string {

	localEntries, err := local.Entries(ctx)

	if err != nil {

		panic(err)

	}

	remoteEntries, err := remote.Entries(ctx)

	if err != nil {

		panic(err)

	}

	addedFiles := []*File{}

    deletedFiles := []*File{}

    entriesInBothDirs := []string{}

	modifiedFiles := []*File{}

	for _, localEntry := range localEntries {

		remoteIndex := slices.Index(remoteEntries, localEntry)

		if remoteIndex != -1 {

			entriesInBothDirs = append(entriesInBothDirs, localEntry)

            // These two lines delete the element from the slice
            remoteEntries[remoteIndex] = remoteEntries[len(remoteEntries) - 1]

            remoteEntries = remoteEntries[:len(remoteEntries) - 1]

		} else {

			addedFiles = append(addedFiles, local.File(localEntry))

		}

	}

    for _, remoteEntry := range remoteEntries {

        deletedFiles = append(deletedFiles, remote.File(remoteEntry))

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

            modifiedFiles = append(modifiedFiles, local.File(entry))

        }

	}

    fmt.Println(" ----- ADDED ----- ")

    PrintFileList(ctx, addedFiles)

    fmt.Println(" ----- DELETED ----- ")

    PrintFileList(ctx, deletedFiles)

    fmt.Println(" ----- MODIFIED ----- ")

    PrintFileList(ctx, modifiedFiles)

	return ""
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
