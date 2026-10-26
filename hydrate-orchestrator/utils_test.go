package main

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestKubernetesAppDeploymentDescribesEveryImageUpdate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		images []ImageData
		want   []string
	}{
		{"services", []ImageData{
			{Platform: "cluster", Tenant: "tenant", Env: "dev", ServiceNameList: []string{"micro-a", "micro-c"}, Image: "image-A", RepositoryCaller: "caller"},
			{Platform: "cluster", Tenant: "tenant", Env: "dev", ServiceNameList: []string{"micro-b"}, Image: "image-B", RepositoryCaller: "caller"},
		}, []string{"Service `micro-a`: `image-A`", "Service `micro-c`: `image-A`", "Service `micro-b`: `image-B`"}},
		{"keys", []ImageData{
			{ImageKeys: []string{"/a/image"}, Image: "image-A"},
			{ImageKeys: []string{"/b/image"}, Image: "image-B"},
		}, []string{"Image key `/a/image`: `image-A`", "Image key `/b/image`: `image-B`"}},
		{"mixed", []ImageData{
			{ServiceNameList: []string{"micro-a"}, Image: "image-A"},
			{ImageKeys: []string{"/b/image"}, Image: "image-B"},
		}, []string{"Service `micro-a`: `image-A`", "Image key `/b/image`: `image-B`"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input, err := json.Marshal(ImageMatrix{Images: tc.images})
			if err != nil {
				t.Fatal(err)
			}
			m := &HydrateOrchestrator{}
			deployments := m.processImagesMatrix(string(input)).KubernetesDeployments
			if len(deployments) != 1 {
				t.Fatalf("expected one grouped deployment, got %d", len(deployments))
			}
			deployment := &deployments[0]
			services := slices.Clone(deployment.ServiceNames)
			body := deployment.String(false, "https://github.com/org/caller")
			title := deployment.String(true)
			if title != deployment.String(true) || !slices.Equal(services, deployment.ServiceNames) {
				t.Fatal("formatting title mutated service names")
			}
			if body != deployment.String(false, "https://github.com/org/caller") {
				t.Fatal("formatting title changed PR body")
			}
			for _, detail := range tc.want {
				if !strings.Contains(body, detail) {
					t.Errorf("missing %q in PR body: %s", detail, body)
				}
			}
		})
	}
}

func TestKubernetesAppDeploymentManualDescription(t *testing.T) {
	deployment := KubernetesAppDeployment{Cluster: "cluster", Tenant: "tenant", Environment: "dev"}
	body := deployment.String(false)
	if strings.Contains(body, "Image updates") || !strings.Contains(body, "Cluster: `cluster`") {
		t.Fatalf("unexpected manual deployment description: %s", body)
	}
}
