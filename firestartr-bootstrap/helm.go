package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *FirestartrBootstrap) BuildHelmValues(ctx context.Context) string {

	helmValues := &HelmValues{
		Labels: Labels{},
		Deploy: Deploy{
			Replicas: 1,
			Image: Image{
				Name: "ghcr.io/prefapp/gitops-k8s",
				Tag: fmt.Sprintf(
					"%s_full-%s",
					m.Bootstrap.Firestartr.Version,
					m.Creds.CloudProvider.Name,
				),
				PullPolicy: "Always",
			},
			Command:       []string{"./run.sh", "operator", "--start", "controller"},
			ContainerPort: 80,
			Probes:        Probes{},
			Resources:     nil,
			VolumeMounts:  nil,
			Volumes:       nil,
		},
		ServiceAccount: ServiceAccount{
			Annotations: nil,
		},
		RoleRules: []RoleRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
		},
		Secret: Secret{
			Type: "Opaque",
			Data: map[string]string{
				"GITHUB_APP_PEM_FILE": m.Creds.GithubApp.Pem,
			},
		},
		Config: Config{
			Data: map[string]string{
				"OPERATOR_KIND_LIST": strings.Join([]string{
					"githubgroups",
					"githubmemberships",
					"githubrepositories",
					"githubrepositoryfeatures",
					"terraformworkspaces",
					"terraformworkspaceplans",
				}, ","),
				"OPERATOR_NAMESPACE":                 "default",
				"OPERATOR_IGNORE_LEASE":              "true",
				"GITHUB_APP_ID":                      m.Creds.GithubApp.GhAppId,
				"GITHUB_APP_INSTALLATION_ID":         m.Creds.GithubApp.InstallationId,
				"GITHUB_APP_INSTALLATION_ID_PREFAPP": m.Creds.GithubApp.PrefappInstallationId,
				"NODE_TLS_REJECT_UNAUTHORIZED":       "0",
				"ORG":                                m.GhOrg,
				"DEBUG":                              "*",
			},
		},
	}

	return dumpValuesToYaml(ctx, helmValues)
}

func dumpValuesToYaml(

	ctx context.Context,

	values *HelmValues,

) string {

	yamlContent, err := yaml.Marshal(values)

	if err != nil {

		panic(err)

	}

	return string(yamlContent)

}

func encodeB64DaggerSecret(ctx context.Context, text string) string {

	return base64.StdEncoding.EncodeToString([]byte(text))

}
