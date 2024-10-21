package main

import (
	"context"
	"dagger/notify-and-hydrate-state/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"

	"reflect"

	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/yaml"
)

func (m *NotifyAndHydrateState) CompareDirs(

	ctx context.Context,

	oldCrs *dagger.Directory,

	newCrs *dagger.Directory,

	affectedClaims []string,

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

			if m.IsAffectedCRFromPr(ctx, affectedClaims, newCrs.File(newEntry)) {

				entriesInBothDirs = append(entriesInBothDirs, newEntry)

				// These two lines delete the element from the slice
				oldEntries[oldIndex] = oldEntries[len(oldEntries)-1]

				oldEntries = oldEntries[:len(oldEntries)-1]
			}

		} else {

			if m.IsAffectedCRFromPr(ctx, affectedClaims, newCrs.File(newEntry)) {

				result.AddedFiles = append(result.AddedFiles, newCrs.File(newEntry))

			}
		}

	}

	for _, oldEntry := range oldEntries {

		if m.IsAffectedCRFromPr(ctx, affectedClaims, oldCrs.File(oldEntry)) {

			result.DeletedFiles = append(result.DeletedFiles, oldCrs.File(oldEntry))

		}
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

			if m.IsAffectedCRFromPr(ctx, affectedClaims, newCrs.File(entry)) {

				result.ModifiedFiles = append(result.ModifiedFiles, newCrs.File(entry))

			}
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

	listToPrint []*dagger.File,

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

	ignoredAnnotations := []string{
		"firestartr.dev/last-state-pr",
		"firestartr.dev/last-claim-pr",
	}

	var obj1, obj2 map[string]interface{}

	json.Unmarshal([]byte(jsonString1), &obj1)

	json.Unmarshal([]byte(jsonString2), &obj2)

	for _, annotation := range ignoredAnnotations {

		obj1["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[annotation] = nil

		obj2["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})[annotation] = nil

	}

	return reflect.DeepEqual(obj1, obj2)
}

func (m *NotifyAndHydrateState) IsAffectedCRFromPr(

	ctx context.Context,

	affectedClaims []string,

	cr *dagger.File,

) bool {

	contents, err := cr.Contents(ctx)

	if err != nil {

		panic(err)

	}

	jsonContents, err := yaml.YAMLToJSON([]byte(contents))

	if err != nil {

		panic(err)

	}

	annotations := gjson.Get(string(jsonContents), "metadata.annotations")

	claimRef := annotations.Get(gjson.Escape("firestartr.dev/claim-ref")).String()

	claimName := strings.Split(claimRef, "/")[1]

	isAffected := slices.Contains(affectedClaims, claimName)

	return isAffected

}
