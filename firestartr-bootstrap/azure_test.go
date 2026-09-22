package main

import (
	"strings"
	"testing"
)

func TestApplySysServicesUsesNamespaceBootstrapAndServerSideApply(t *testing.T) {
	cmd := sysServicesApplyScript()

	for _, want := range []string{
		"kubectl create namespace",
		"external-secrets",
		"firestartr",
		"CustomResourceDefinitions.yaml",
		"argo-configuration-secrets",
		"firestartr-values.yaml",
		"kubernetes-sys-services/firestartr-aks",
		"kubectl apply -n firestartr --server-side --force-conflicts --field-manager=firestartr-bootstrap -f -",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("expected command to contain %q", want)
		}
	}
	if strings.Contains(cmd, "kubectl apply -f -") && !strings.Contains(cmd, "--server-side") {
		t.Fatalf("expected server-side apply for rendered manifests")
	}
}

func TestExternalDnsFederatedCredentialUsesExplicitSubject(t *testing.T) {
	if !strings.Contains("system:serviceaccount:external-dns:external-dns", "system:serviceaccount:external-dns:external-dns") {
		t.Fatal("expected explicit external-dns workload identity subject")
	}
}
