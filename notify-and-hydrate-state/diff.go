package main

import (
	"context"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/yaml"
)

type DiffResult struct {
	AddedFiles []*File

	DeletedFiles []*File

	ModifiedFiles []*File

	UnmodifiedFiles []*File
}

func (m *NotifyAndHydrateState) CompareDirs(

	ctx context.Context,

	oldCrs *Directory,

	newCrs *Directory,

) DiffResult {

	oldEntries, err := oldCrs.
		WithoutDirectory(".config").
		WithoutDirectory(".git").
		WithoutDirectory(".github").
		Glob(ctx, "*.yaml")

	if err != nil {

		panic(err)

	}

	newEntries, err := newCrs.Entries(ctx)

	if err != nil {

		panic(err)

	}

	result := DiffResult{}

	entriesInBothDirs := []string{}

	for _, newEntry := range newEntries {

		oldIndex := slices.Index(oldEntries, newEntry)

		if oldIndex != -1 {

			entriesInBothDirs = append(entriesInBothDirs, newEntry)

			// These two lines delete the element from the slice
			oldEntries[oldIndex] = oldEntries[len(oldEntries)-1]

			oldEntries = oldEntries[:len(oldEntries)-1]

		} else {

			result.AddedFiles = append(result.AddedFiles, newCrs.File(newEntry))

		}

	}

	for _, oldEntry := range oldEntries {

		result.DeletedFiles = append(result.DeletedFiles, oldCrs.File(oldEntry))

	}

	for _, entry := range entriesInBothDirs {

		oldContents, err := oldCrs.File(entry).Contents(ctx)

		if err != nil {

			panic(err)

		}

		newContents, err := newCrs.File(entry).Contents(ctx)

		if err != nil {

			panic(err)

		}

		if !m.AreYamlsEqual(ctx, oldContents, newContents) {

			result.ModifiedFiles = append(result.ModifiedFiles, newCrs.File(entry))

		} else {

			result.UnmodifiedFiles = append(result.UnmodifiedFiles, newCrs.File(entry))
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

func (m *NotifyAndHydrateState) AreYamlsEqual(

	ctx context.Context,

	yamlA string,

	yamlB string,

) bool {

	jsonString1, err := yaml.YAMLToJSON([]byte(yamlA))

	jsonString2, err2 := yaml.YAMLToJSON([]byte(yamlB))

	if err != nil {

		panic(err)

	}

	if err2 != nil {

		panic(err2)

	}

	return jsonpatch.Equal(jsonString1, jsonString2)

}
