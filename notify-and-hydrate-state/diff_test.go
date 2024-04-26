package main

import (
	"context"
	"testing"
)

func TestDiff(t *testing.T) {

	content1 :=
		`apiVersion: firestartr.dev/v1
kind: FirestartrTerraformWorkspace
metadata:
  annotations:
    firestartr.dev/claim-ref: TFWorkspaceClaim/test-module-a
    firestartr.dev/external-name: test_module_a
    firestartr.dev/policy: apply
  labels:
    claim-ref: test-module-a
  name: test-module-a-3ff33491-57e2-47cb-89ec-1c1cdcc65a4b
spec:
  context:
    backend:
      ref:
        kind: FirestartrProviderConfig
        name: firestartr-terraform-state
    providers:
      - ref:
          kind: FirestartrProviderConfig
          name: provider-aws-workspaces
  firestartr:
    tfStateKey: 3ff33491-57e2-47cb-89ec-1c1cdcc65a4b
  module: |
    output "hello" {
      value = "Hello, World!"
    }
  source: Inline
  values: "{}"
  references: []`

	content2 :=
`apiVersion: firestartr.dev/v1
kind: FirestartrTerraformWorkspace
metadata:
  annotations:
    firestartr.dev/claim-ref: TFWorkspaceClaim/test-module-a
    firestartr.dev/external-name: test_module_a
    firestartr.dev/policy: apply-a
  labels:
    claim-ref: test-module-a
  name: test-module-a-3ff33491-57e2-47cb-89ec-1c1cdcc65a4b
spec:
  context:
    backend:
      ref:
        kind: FirestartrProviderConfig
        name: firestartr-terraform-state
    providers:
      - ref:
          kind: FirestartrProviderConfig
          name: provider-aws-workspaces
  firestartr:
    tfStateKey: 3ff33491-57e2-47cb-89ec-1c1cdcc65a4b
  module: |
    output "hello" {
      value = "Hello, World!"
    }
  source: Inline
  values: "{}"
  references: []`		

	m := NotifyAndHydrateState{}

	if m.AreYamlsEqual(context.Background(), content1, content2) {

		t.Errorf("Yamls are equal, expected different")

	}

}
