package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestCanPatches(t *testing.T) {

	// ctx := context.Background()

	m := &HydrateKubernetes{}

	completePath := "/micro-a/image/a/b/c"

	splitted := strings.Split(completePath, "/")

	fmt.Printf("splitted: %v\n", splitted)

	splittedWithoutLast := splitted[:len(splitted)-1]

	jsonObj := "{}"

	path := ""

	for _, s := range splittedWithoutLast {

		if s == "" {

			continue

		}

		path = path + "/" + s

		patch := fmt.Sprintf("[{\"op\": \"add\", \"path\": \"%s\", \"value\": {}}]", path)

		jsonObj = m.ApplyPatch(jsonObj, patch)

	}

	patch := fmt.Sprintf("[{\"op\": \"add\", \"path\": \"%s\", \"value\": \"my-image:latest\"}]", completePath)

	jsonObj = m.ApplyPatch(jsonObj, patch)

	fmt.Printf("jsonObj: %v\n", jsonObj)
}
