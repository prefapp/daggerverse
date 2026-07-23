package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const SECRETS_FILE_PATH = "/secret_store/secrets.yaml"
const BOOTSTRAP_SECRETS_FILE_PATH = "/secret_store/bootstrap_secrets.yaml"
const OPERATOR_SECRETS_FILE_PATH = "/secret_store/operator_secrets.yaml"

// Maps can't be actual constants in Go, so we use a variable here.
var CREDS_SECRET_LIST = map[string]string{
	"GhAppId":        "ref:secretsclaim:bootstrap-secrets:fs-admin-appid",
	"InstallationId": "ref:secretsclaim:bootstrap-secrets:fs-admin-installationid",
	"Pem":            "ref:secretsclaim:bootstrap-secrets:fs-admin-pem",
}

var OPERATOR_CREDS_SECRET_LIST = map[string]string{
	"GhAppId":        "ref:secretsclaim:operator-secrets:fs-appid",
	"InstallationId": "ref:secretsclaim:operator-secrets:fs-installationid",
	"Pem":            "ref:secretsclaim:operator-secrets:fs-pem",
}

// isAzureProvider returns true when the cloud provider is Azure (azure or azurerm).
func isAzureProvider(providerName string) bool {
	return providerName == "azure" || providerName == "azurerm"
}

func (m *FirestartrBootstrap) CreateKubernetesSecrets(
	ctx context.Context,
	kindContainer *dagger.Container,
) (*dagger.Container, error) {
	// Select the credentials secret template based on the cloud provider.
	secretTemplatePath := "templates/secret.tmpl"
	if isAzureProvider(m.Creds.CloudProvider.Name) {
		secretTemplatePath = "templates/azure_secret.tmpl"
	}

	secretsTmpl, err := dag.CurrentModule().
		Source().
		File(secretTemplatePath).
		Contents(ctx)
	if err != nil {
		return nil, err
	}

	secretsCr, err := renderTmpl(secretsTmpl, m.Creds)
	if err != nil {
		return nil, err
	}

	// Select ExternalSecret templates based on the cloud provider.
	bootstrapSecretsTmplPath := "external_secrets/bootstrap_secrets.tmpl"
	operatorSecretsTmplPath := "external_secrets/operator_secrets.tmpl"
	if isAzureProvider(m.Creds.CloudProvider.Name) {
		bootstrapSecretsTmplPath = "external_secrets/azure_bootstrap_secrets.tmpl"
		operatorSecretsTmplPath = "external_secrets/azure_operator_secrets.tmpl"
	}

	bootstrapSecretsTmpl, err := dag.CurrentModule().
		Source().
		File(bootstrapSecretsTmplPath).
		Contents(ctx)
	if err != nil {
		return nil, err
	}

	bootstrapSecretsCr, err := renderTmpl(bootstrapSecretsTmpl, m)
	if err != nil {
		return nil, err
	}

	operatorSecretsTmpl, err := dag.CurrentModule().
		Source().
		File(operatorSecretsTmplPath).
		Contents(ctx)
	if err != nil {
		return nil, err
	}

	operatorSecretsCr, err := renderTmpl(operatorSecretsTmpl, m)
	if err != nil {
		return nil, err
	}

	// Render and mount the SecretStore manifest based on the cloud provider.
	const secretStorePath = "/secret_store/secretstore.yaml"
	var secretStoreCr string
	if isAzureProvider(m.Creds.CloudProvider.Name) {
		azureSecretStoreTmpl, err := dag.CurrentModule().
			Source().
			File("external_secrets/azure_secretstore.tmpl").
			Contents(ctx)
		if err != nil {
			return nil, err
		}
		secretStoreCr, err = renderTmpl(azureSecretStoreTmpl, m.Creds)
		if err != nil {
			return nil, err
		}
	} else {
		awsSecretStoreContent, err := dag.CurrentModule().
			Source().
			File("external_secrets/aws_secretstore.yaml").
			Contents(ctx)
		if err != nil {
			return nil, err
		}
		secretStoreCr = awsSecretStoreContent
	}

	firestartrPodName, err := kindContainer.
		WithExec([]string{
			"kubectl", "get", "pod",
			"-l", "app.kubernetes.io/name=external-secrets-webhook",
			"-o", "name",
			"-n", "external-secrets",
		}).
		Stdout(ctx)
	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to get external-secrets pod")
		return nil, errors.New(errMsg)
	}

	pushSecretsDirectory, err := m.GeneratePushSecrets(ctx)
	if err != nil {
		return nil, err
	}

	kindContainer, err = kindContainer.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithDirectory("/push-secrets", pushSecretsDirectory).
		WithNewFile(SECRETS_FILE_PATH, secretsCr).
		WithNewFile(BOOTSTRAP_SECRETS_FILE_PATH, bootstrapSecretsCr).
		WithNewFile(OPERATOR_SECRETS_FILE_PATH, operatorSecretsCr).
		WithExec([]string{
			"kubectl", "apply", "-f", SECRETS_FILE_PATH,
		}).
		WithExec([]string{
			"kubectl", "wait",
			"--for=condition=Ready",
			strings.Trim(firestartrPodName, "\n"),
			"--timeout=10h",
			"-n", "external-secrets",
		}).
		WithNewFile(secretStorePath, secretStoreCr).
		WithExec([]string{
			"kubectl", "apply", "-f", secretStorePath,
		}).
		WithExec([]string{
			"kubectl", "apply", "-f", BOOTSTRAP_SECRETS_FILE_PATH,
		}).
		WithExec([]string{
			"kubectl",
			"wait",
			"--for=create",
			"--timeout=10h",
			"secret/bootstrap-secrets",
		}).
		WithExec([]string{
			"kubectl", "apply", "-f", OPERATOR_SECRETS_FILE_PATH,
		}).
		WithExec([]string{
			"kubectl",
			"wait",
			"--for=create",
			"--timeout=10h",
			"secret/operator-secrets",
		}).
		WithExec([]string{
			"kubectl", "apply", "-f", "/push-secrets/push-secrets.yaml",
		}).
		WithExec([]string{
			"kubectl",
			"wait",
			"--for=condition=Ready=True",
			"--timeout=10h",
			"pushsecret/webhook-pushsecret",
		}).
		WithExec([]string{
			"kubectl",
			"wait",
			"--for=condition=Ready=True",
			"--timeout=10h",
			"pushsecret/prefapp-bot-pat-pushsecret",
		}).
		Sync(ctx)

	if err != nil {
		errMsg := extractErrorMessage(err, "Failed to create Kubernetes secrets")
		return nil, errors.New(errMsg)
	}

	return kindContainer, nil
}

func (m *FirestartrBootstrap) PopulateGithubAppCredsFromSecrets(
	ctx context.Context,
	kindContainer *dagger.Container,
) error {
	// Get the GitHub App credentials struct
	credsReflector := reflect.ValueOf(&m.Creds.GithubApp).Elem()
	credsReflectorOperator := reflect.ValueOf(&m.Creds.GithubAppOperator).Elem()

	// For each known secret property
	for property, ref := range CREDS_SECRET_LIST {
		// Fetch the secret value from Kubernetes
		secretValue, err := m.GetKubernetesSecretValue(ctx, kindContainer, ref)
		if err != nil {
			return err
		}

		// Check it exists and is settable within the struct
		field := credsReflector.FieldByName(property)
		if !field.IsValid() {
			return fmt.Errorf(
				"Field %q does not exist in GithubApp struct", property,
			)
		}
		if !field.CanSet() {
			return fmt.Errorf(
				"Field %q in GithubApp struct is not settable", property,
			)
		}

		// Set the field to the fetched secret value
		field.SetString(secretValue)
	}

	// For each known secret property (for the operator)
	for property, ref := range OPERATOR_CREDS_SECRET_LIST {
		// Fetch the secret value from Kubernetes
		secretValue, err := m.GetKubernetesSecretValue(ctx, kindContainer, ref)
		if err != nil {
			return err
		}

		// Check it exists and is settable within the struct
		field := credsReflectorOperator.FieldByName(property)
		if !field.IsValid() {
			return fmt.Errorf(
				"Field %q does not exist in Operator's GithubApp struct", property,
			)
		}
		if !field.CanSet() {
			return fmt.Errorf(
				"Field %q in Operator's GithubApp struct is not settable", property,
			)
		}

		// Set the field to the fetched secret value
		field.SetString(secretValue)
	}

	// Lastly, set the RawPem field to the escaped version of the Pem field
	// for use in the ProviderConfig template
	escaped := strings.ReplaceAll(m.Creds.GithubApp.Pem, "\n", "\\n")
	m.Creds.GithubApp.RawPem = escaped

	escaped = strings.ReplaceAll(m.Creds.GithubAppOperator.Pem, "\n", "\\n")
	m.Creds.GithubAppOperator.RawPem = escaped

	return nil
}

func (m *FirestartrBootstrap) GetKubernetesSecretValue(
	ctx context.Context,
	kindContainer *dagger.Container,
	fullRef string,
) (string, error) {
	secretRef := strings.Replace(fullRef, "ref:secretsclaim:", "", 1)
	secretCR := strings.Split(secretRef, ":")[0]
	secretName := strings.Split(secretRef, ":")[1]

	encodedValue, err := kindContainer.
		WithExec([]string{
			"kubectl", "get", "secret", secretCR,
			"-o", fmt.Sprintf("jsonpath=\"{.data.%s}\"", secretName),
		}).Stdout(ctx)
	if err != nil {
		errMsg := extractErrorMessage(
			err, fmt.Sprintf("Failed to get Kubernetes secret %s/%s", secretCR, secretName),
		)
		return "", errors.New(errMsg)
	}

	return kindContainer.
		WithNewFile("/tmp/encoded_value.txt", strings.Trim(encodedValue, "\"\n")).
		WithExec([]string{
			"base64", "-d", "/tmp/encoded_value.txt",
		}).Stdout(ctx)

}
