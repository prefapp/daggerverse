package main

import (
	"context"
	"fmt"

	"dagger/firestartr-bootstrap/internal/dagger"
)

func (m *FirestartrBootstrap) RenderDeployment(
	ctx context.Context,
) (*dagger.Directory, error) {

    accountID, err := m.ValidateSTSCredentials(ctx)

    if err != nil {
        return nil, fmt.Errorf("Obtaining the accountID of aws: %s", err)
    }


	// let's populate the struct
	deploymentData := DeploymentConfig{

		Org:		m.Bootstrap.Org,
		Customer:	m.Bootstrap.Customer,
		Webhook:	DeploymentWebhook {

			URL:		m.Bootstrap.WebhookUrl,
			Secret: 	m.Bootstrap.WebhookSecretRef,

		},

		ExternalSecrets: DeploymentExternalSecrets{

            RoleARN:	fmt.Sprintf("arn:aws:iam::%s:role/Firestartr-%s", 

                accountID,

                m.Bootstrap.Customer,

            ),
			
		},

		Controller: DeploymentController{

			Image: fmt.Sprintf("ghcr.io/prefapp/gitops-k8s:%s", fmt.Sprintf(

				"%s_full-%s",
				m.Bootstrap.Firestartr.OperatorVersion,
				m.Creds.CloudProvider.Name,

			)),

			RoleARN:	"controller-role-ref",

            GithubApp: 	DeploymentGithubApp{

                GithubAppId: 	"ref to id",
                GithubAppPem: 	"ref to pem",

            },


		},

		Aws: 	DeploymentAws{
	
			Bucket:				*m.Creds.CloudProvider.Config.Bucket,
			Region:				m.Creds.CloudProvider.Config.Region,

		},

		Provider:  DeploymentGithubApp{

			GithubAppId: 		"ref to id",
			GithubAppPem: 	"ref to pem",

		},
	}

	deploymentTemplateFile := dag.CurrentModule().
		Source().
		File("templates/deployment/values.tmpl")

	templateContent, err := deploymentTemplateFile.Contents(ctx)
	if err != nil {
		return nil, err
	}

	rendered, err := renderTmpl(templateContent, deploymentData)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("values.yaml", rendered), nil

}

