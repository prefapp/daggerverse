package main

import (
    "context"
	"dagger/firestartr-bootstrap/internal/dagger"
)

func (m *FirestartrBootstrap) GeneratePushSecrets(
    ctx context.Context,
) (*dagger.Directory, error) {

   webHookPushSecret := PushSecretElement{
       Name: "webhook-pushsecret",
       KubernetesSecret: "webhook-secret",
       KubernetesSecretKey: "webhook-secret-key",
       ParameterName: m.Bootstrap.WebhookSecretRef,
       Value: "my-secret-secret",
       SecretStore: "aws",
   }

   prefappBotPatSecret := PushSecretElement{
       Name: "prefapp-bot-pat-pushsecret",
       KubernetesSecret: "prefapp-bot-secret",
       KubernetesSecretKey: "botpat-secret-key",
       ParameterName: m.Bootstrap.PrefappBotPatSecretRef,
       Value: m.Creds.GithubApp.PrefappBotPat,
       SecretStore: "aws",

   }

   prefappCliVersion := PushSecretElement{
       Name: "prefapp-bot-pat-pushsecret",
       KubernetesSecret: "prefapp-cli-version",
       KubernetesSecretKey: "cli-version-key",
       ParameterName: m.Bootstrap.FirestartrCliVersionSecretRef,
       Value: m.Bootstrap.Firestartr.CliVersion,
       SecretStore: "aws",

   }

   rendered, err := renderPushSecret(ctx, &webHookPushSecret, "external_secrets/push_secret.tmpl")

   if err != nil {
        return nil, err
   }

   renderedSecret, err := renderPushSecret(ctx, &webHookPushSecret, "external_secrets/secret.tmpl")

   if err != nil {
        return nil, err
   }


   renderedGH, err := renderPushSecret(ctx, &prefappBotPatSecret, "external_secrets/push_secret.tmpl")

   if err != nil {
        return nil, err
   }

   renderedSecretGH, err := renderPushSecret(ctx, &prefappBotPatSecret, "external_secrets/secret.tmpl")

   if err != nil {
        return nil, err
   }

   renderedCli, err := renderPushSecret(ctx, &prefappCliVersion, "external_secrets/push_secret.tmpl")

   if err != nil {
        return nil, err
   }

   renderedSecretCli, err := renderPushSecret(ctx, &prefappCliVersion, "external_secrets/secret.tmpl")

   if err != nil {
        return nil, err
   }


   return dag.Directory().WithNewFile("push-secrets.yaml", rendered + renderedSecret + renderedGH + renderedSecretGH + renderedCli + renderedSecretCli), nil
}

func renderPushSecret(
    ctx context.Context, 
    data *PushSecretElement,
    template string,
) (string, error){

	psTemplateFile := dag.CurrentModule().
		Source().
		File(template)

	templateContent, err := psTemplateFile.Contents(ctx)
	if err != nil {
		return "", err
	}

	rendered, err := renderTmpl(templateContent, data)
	if err != nil {
		return "", err
	}

    return rendered, nil
}
