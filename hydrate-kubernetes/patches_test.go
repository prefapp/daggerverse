package main

import (
	"testing"
)

func TestCanPatches(t *testing.T) {

	m := &HydrateKubernetes{}

	completePath := "/a/b/c/d/e"

	res := m.GenerateValue(completePath, "test")

	if res != "{\"a\":{\"b\":{\"c\":{\"d\":{\"e\":\"test\"}}}}}" {

		t.Errorf("Expected {\"a\":{\"b\":{\"c\":{\"d\":{\"e\":\"test\"}}}}}, got %v", res)

	}
}
