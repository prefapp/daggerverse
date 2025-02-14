package main

import (
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
)

func (m *HydrateKubernetes) ApplyPatch(document string, patch string) string {

	patchDecoded, err := jsonpatch.DecodePatch([]byte(patch))

	if err != nil {
		panic(err)
	}

	modified, err := patchDecoded.Apply([]byte(document))

	if err != nil {
		panic(err)
	}

	return string(modified)
}

func (m *HydrateKubernetes) GenerateValue(fullPath string, value string) string {

	splitted := strings.Split(fullPath, "/")

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

	patch := fmt.Sprintf("[{\"op\": \"add\", \"path\": \"%s\", \"value\": \"%s\"}]", fullPath, value)

	return m.ApplyPatch(jsonObj, patch)

}
