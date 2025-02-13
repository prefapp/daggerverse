package main

import (
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
