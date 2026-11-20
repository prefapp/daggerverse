package main

import "testing"

// TestInjectDedicatedComponentsDoesNotDuplicateFeatureAtDifferentVersion ensures
// that an existing state_repo_sys_services feature at a version other than the
// default is preserved instead of producing a second, conflicting entry.
func TestInjectDedicatedComponentsDoesNotDuplicateFeatureAtDifferentVersion(t *testing.T) {
	bootstrap := &Bootstrap{
		Components: []Component{
			{
				Name: "state-sys-services",
				Features: []Feature{
					{Name: "state_repo_sys_services", Version: "2.5.0"},
				},
			},
			{Name: "state-argocd"},
		},
	}

	components := injectDedicatedComponents(bootstrap)

	var sysSvc *Component
	for i := range components {
		if components[i].Name == "state-sys-services" {
			sysSvc = &components[i]
		}
	}
	if sysSvc == nil {
		t.Fatal("expected state-sys-services component to be present")
	}

	matching := 0
	for _, feature := range sysSvc.Features {
		if feature.Name == "state_repo_sys_services" {
			matching++
			if feature.Version != "2.5.0" {
				t.Errorf("state_repo_sys_services version = %q, want preserved %q", feature.Version, "2.5.0")
			}
		}
	}
	if matching != 1 {
		t.Fatalf("expected exactly one state_repo_sys_services feature entry, got %d", matching)
	}
}

// TestInjectDedicatedComponentsAddsDefaultFeatureWhenMissing covers the common
// case where the feature is entirely absent and must be added with the
// default version.
func TestInjectDedicatedComponentsAddsDefaultFeatureWhenMissing(t *testing.T) {
	bootstrap := &Bootstrap{
		Components: []Component{
			{Name: "state-sys-services"},
		},
	}

	components := injectDedicatedComponents(bootstrap)

	var sysSvc *Component
	for i := range components {
		if components[i].Name == "state-sys-services" {
			sysSvc = &components[i]
		}
	}
	if sysSvc == nil {
		t.Fatal("expected state-sys-services component to be present")
	}

	matching := 0
	for _, feature := range sysSvc.Features {
		if feature.Name == "state_repo_sys_services" {
			matching++
			if feature.Version != "2.4.0" {
				t.Errorf("state_repo_sys_services version = %q, want default %q", feature.Version, "2.4.0")
			}
		}
	}
	if matching != 1 {
		t.Fatalf("expected exactly one state_repo_sys_services feature entry, got %d", matching)
	}
}
