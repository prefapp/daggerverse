package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestInitialClaimsBootstrapSecretStore(t *testing.T) {
	content, err := os.ReadFile("templates/initial_claims.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		mode       string
		finalStore string
		want       string
	}{
		{"default", "", "customer-aws-parameter-store", "aws"},
		{"saas", DeploymentModeSaaS, "customer-aws-parameter-store", "aws"},
		{"dedicated", DeploymentModeDedicated, "firestartr-kv", "firestartr-kv"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rendered, err := renderTmpl(string(content), Bootstrap{
				DeploymentMode:         tc.mode,
				FinalSecretStoreName:   tc.finalStore,
				Org:                    "example-org",
				Customer:               "customer",
				DefaultFirestartrGroup: "firestartr",
				DefaultGroup:           "developers",
				DefaultSystemName:      "default-system",
				DefaultDomainName:      "default-domain",
			})
			if err != nil {
				t.Fatal(err)
			}
			decoder := yaml.NewDecoder(strings.NewReader(rendered))
			for {
				var claim struct {
					Kind      string `yaml:"kind"`
					Providers struct {
						ExternalSecrets struct {
							SecretStore struct {
								Name string `yaml:"name"`
							} `yaml:"secretStore"`
						} `yaml:"external_secrets"`
					} `yaml:"providers"`
				}
				if err := decoder.Decode(&claim); err == io.EOF {
					t.Fatal("SecretsClaim not found")
				} else if err != nil {
					t.Fatal(err)
				}
				if claim.Kind == "SecretsClaim" {
					if got := claim.Providers.ExternalSecrets.SecretStore.Name; got != tc.want {
						t.Fatalf("bootstrap SecretStore = %q, want %q", got, tc.want)
					}
					break
				}
			}
		})
	}
}
