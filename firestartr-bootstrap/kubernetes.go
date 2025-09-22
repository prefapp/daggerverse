package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"strings"
	"time"
)

const SECRETS_FILE_PATH = "/secret_store/secrets.yaml"

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
		WithExec([]string{
			"kubectl", "apply", "-f", SECRETS_FILE_PATH,
		}).
		WithExec([]string{
			"kubectl", "wait",
			"--for=condition=Ready",
			strings.Trim(firestartrPodName, "\n"),
			"--timeout=300s",
			"-n", "external-secrets",
		}).
		WithFile("/secret_store/aws_secretstore.yaml", awsSecretStoreFile).
		WithExec([]string{
			"kubectl", "apply", "-f", "/secret_store/aws_secretstore.yaml",
		}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	return kindContainer, nil
}
