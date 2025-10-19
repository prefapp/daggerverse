package main

import (
	"testing"
)

func TestGenerateOjectFromPath(t *testing.T) {

	m := &HydrateKubernetes{}

	fullPath := "/a/b/c/d/e"

	res := m.GenerateOjectFromPath(fullPath, "test", "{}")

	if res != "{\"a\":{\"b\":{\"c\":{\"d\":{\"e\":\"test\"}}}}}" {

		t.Errorf("Expected {\"a\":{\"b\":{\"c\":{\"d\":{\"e\":\"test\"}}}}}, got %v", res)

	}

}
