package main

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"

	sigsyaml "sigs.k8s.io/yaml"
)

func (m *HydrateSecrets) PatchClaim(

	path string,

	value string,

	yamlContent string,

) (string, error) {

	tojson, err := sigsyaml.YAMLToJSON([]byte(yamlContent))
	if err != nil {
		return "", err
	}

	patchJSON := []byte(fmt.Sprintf(
		`[{"op": "add", "path": "%s", "value": %s}]`,
		path,
		value,
	))

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		panic(err)
	}

	modifiedJson, err := patch.Apply([]byte(tojson))
	if err != nil {
		return "", err
	}

	modifiedYaml, err := sigsyaml.JSONToYAML(modifiedJson)
	if err != nil {
		return "", err
	}

	return string(modifiedYaml), nil
}
