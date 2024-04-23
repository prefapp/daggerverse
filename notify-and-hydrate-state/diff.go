package main

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"
)

func (m *NotifyAndHydrateState) CompareDirs(

	ctx context.Context,

	dirA Directory,

	dirB Directory,

) string {

	aEntries, err := dirA.Entries(ctx)

	if err != nil {

		panic(err)

	}

	bEntries, err := dirB.Entries(ctx)

	if err != nil {

		panic(err)

	}

	files := []*File{}

	for _, aEntry := range aEntries {

		isContained := slices.Contains(bEntries, aEntry)

		if !isContained {

			files = append(files, dirA.File(aEntry))

		} else {

			contentsFromDirA, err := dirA.File(aEntry).Contents(ctx)

			if err != nil {

				panic(err)

			}

			contentsFromDirB, err := dirB.File(aEntry).Contents(ctx)

			if err != nil {

				panic(err)

			}

			if contentsFromDirA != contentsFromDirB {

				files = append(files, dirA.File(aEntry))

			}

		}

	}

	for _, bEntry := range bEntries {

		isContained := slices.Contains(aEntries, bEntry)

		if !isContained {

			files = append(files, dirB.File(bEntry))

		} else {

			contentsFromDirA, err := dirA.File(bEntry).Contents(ctx)

			if err != nil {

				panic(err)

			}

			contentsFromDirB, err := dirB.File(bEntry).Contents(ctx)

			if err != nil {

				panic(err)

			}

			if contentsFromDirA != contentsFromDirB {

				files = append(files, dirB.File(bEntry))

			}

		}

	}

	for _, file := range files {

		contents, err := file.Contents(ctx)

		if err != nil {

			panic(err)

		}

		fmt.Print(contents)

	}

	return ""
}
