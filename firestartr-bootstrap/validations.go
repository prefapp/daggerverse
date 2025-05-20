package main

import (
	"context"
	"dagger/firestartr-bootstrap/internal/dagger"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

func (m *FirestartrBootstrap) ValidateBootstrapFile(ctx context.Context, bootstrapFile *dagger.File) error {
	schema, err := dag.CurrentModule().Source().File("schemas/bootstrap-file.json").Contents(ctx)
	if err != nil {
		return err
	}

	bootstrapFileContents, err := bootstrapFile.Contents(ctx)
	if err != nil {
		return err
	}

	json, err := yaml.YAMLToJSON([]byte(bootstrapFileContents))
	if err != nil {
		return err
	}

	if err := validateDocumentSchema(schema, string(json)); err != nil {
		return fmt.Errorf("failed to validate bootstrap file: %w", err)
	}
	return nil
}

func (m *FirestartrBootstrap) ValidateCredentialsFile(ctx context.Context, credentialsFileContents string) error {
	schema, err := dag.CurrentModule().Source().File("schemas/credentials-file.json").Contents(ctx)
	if err != nil {
		return err
	}

	jsonDoc, err := yaml.YAMLToJSON([]byte(credentialsFileContents))
	if err != nil {
		return err
	}

	if err := validateDocumentSchema(string(jsonDoc), schema); err != nil {
		return fmt.Errorf("failed to validate credentials file: %w", err)
	}
	return nil
}

func validateDocumentSchema(document string, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(document)

	res, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		return fmt.Errorf("document is not valid %s", res.Errors())
	}
	return nil
}
