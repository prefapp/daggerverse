package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const SECRETS_FILE_PATH = "/secret_store/secrets.yaml"
const BOOTSTRAP_SECRETS_FILE_PATH = "/secret_store/bootstrap_secrets.yaml"

// Maps can't be actual constants in Go, so we use a variable here.
var CREDS_SECRET_LIST = map[string]string{
	"GhAppId":        "ref:secretsclaim:bootstrap-secrets:fs-admin-appid",
	"InstallationId": "ref:secretsclaim:bootstrap-secrets:fs-admin-installationid",
	"Pem":            "ref:secretsclaim:bootstrap-secrets:fs-admin-pem",
	"BotPat":         "ref:secretsclaim:bootstrap-secrets:prefapp-bot-pat",
}

func (m *FirestartrBootstrap) CreateKubernetesSecrets(
	ctx context.Context,
	kindContainer *dagger.Container,
) (*dagger.Container, error) {
	secretsTmpl, err := dag.CurrentModule().
		Source().
		File("templates/secret.tmpl").
		Contents(ctx)

	secretsCr, err := renderTmpl(secretsTmpl, m.Creds)
	if err != nil {
		return nil, err
	}

	bootstrapSecretsTmpl, err := dag.CurrentModule().
		Source().
		File("external_secrets/bootstrap_secrets.tmpl").
		Contents(ctx)

	bootstrapSecretsCr, err := renderTmpl(bootstrapSecretsTmpl, m.Bootstrap)
	if err != nil {
		return nil, err
	}

	awsSecretStoreFile := dag.CurrentModule().
		Source().
		File("external_secrets/aws_secretstore.yaml")

	firestartrPodName, err := kindContainer.
		WithExec([]string{
			"kubectl", "get", "pod",
			"-l", "app.kubernetes.io/name=external-secrets-webhook",
			"-o", "name",
			"-n", "external-secrets",
		}).
		Stdout(ctx)
	if err != nil {
		return nil, err
	}

	kindContainer, err = kindContainer.
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithNewFile(SECRETS_FILE_PATH, secretsCr).
		WithNewFile(BOOTSTRAP_SECRETS_FILE_PATH, bootstrapSecretsCr).
		WithExec([]string{
			"kubectl", "apply", "-f", SECRETS_FILE_PATH,
		}).
		WithExec([]string{
			"kubectl", "wait",
			"--for=condition=Ready",
			strings.Trim(firestartrPodName, "\n"),
			"--timeout=180s",
			"-n", "external-secrets",
		}).
		WithFile("/secret_store/aws_secretstore.yaml", awsSecretStoreFile).
		WithExec([]string{
			"kubectl", "apply", "-f", "/secret_store/aws_secretstore.yaml",
		}).
		WithExec([]string{
			"kubectl", "apply", "-f", BOOTSTRAP_SECRETS_FILE_PATH,
		}).
		WithExec([]string{
			"kubectl",
			"wait",
			"--for=create",
			"secret/bootstrap-secrets",
			"--timeout=60s",
		}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	return kindContainer, nil
}

func (m *FirestartrBootstrap) PopulateCredsFromParameterStore(
	ctx context.Context,
	kindContainer *dagger.Container,
) {
	credsReflector := reflect.ValueOf(&m.Creds.GithubApp).Elem()
	for property, ref := range CREDS_SECRET_LIST {
		secretValue, err := m.GetKubernetesSecretValue(ctx, kindContainer, ref)
		if err != nil {
			panic(err)
		}

		credsReflector.FieldByName(property).SetString(secretValue)
	}

	escaped := strings.ReplaceAll(m.Creds.GithubApp.Pem, "\n", "\\n")

	m.Creds.GithubApp.RawPem = escaped
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
		return "", err
	}

	return kindContainer.
		WithNewFile("/tmp/encoded_value.txt", strings.Trim(encodedValue, "\"\n")).
		WithExec([]string{
			"base64", "-d", "/tmp/encoded_value.txt",
		}).Stdout(ctx)

}
