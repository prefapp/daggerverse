package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (m *FirestartrBootstrap) ValidateSTSCredentials(
	ctx context.Context,
) (string, error) {
	log.Println("Attempting to validate credentials via STS:GetCallerIdentity...")

	cfg := loginAWS(ctx, m.Creds)

	stsClient := sts.NewFromConfig(cfg)

	output, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("authentication failed for STS:GetCallerIdentity: %w", err)
	}

	accountID := aws.ToString(output.Account)

	// 5. Success!
	log.Printf("✅ Credentials validated successfully.")
	log.Printf("   User ARN: %s", aws.ToString(output.Arn))
	log.Printf("   Account ID: %s", aws.ToString(output.Account))

	return accountID, nil
}

func (m *FirestartrBootstrap) ValidateBucket(
	ctx context.Context,
) error {

	cfg := loginAWS(ctx, m.Creds)

	s3Client := s3.NewFromConfig(cfg)

	bucketName := *m.Creds.CloudProvider.Config.Bucket

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err := s3Client.HeadBucket(ctx, input)

	if err == nil {
		log.Printf("✅ Bucket '%s' exists.", bucketName)
		return nil
	}

	return fmt.Errorf("Bucket '%s' does no exist or insufficient permissions", bucketName)

}

func (m *FirestartrBootstrap) ValidateParameters(
	ctx context.Context,
	path string,

) error {

	cfg := loginAWS(ctx, m.Creds)

	// Map the required keys for quick lookup
	requiredMap := make(map[string]struct{})
	for _, key := range m.ExpectedAWSParameters {
		requiredMap[key] = struct{}{}
	}

	// Create a copy of the required map to track which ones are found
	keysToFind := make(map[string]struct{})
	for k, v := range requiredMap {
		keysToFind[k] = v
	}

	ssmClient := ssm.NewFromConfig(cfg)

	noDecrypt := false

	recursive := true

	// Use pagination in case the number of keys exceeds 10 (the default limit)
	paginator := ssm.NewGetParametersByPathPaginator(ssmClient, &ssm.GetParametersByPathInput{
		Path:           &path,
		Recursive:      &recursive,
		WithDecryption: &noDecrypt,
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get parameters from path %s: %w", path, err)
		}

		for _, param := range output.Parameters {
			// Check if the current parameter's Name is one of the keys we are looking for
			if _, ok := keysToFind[*param.Name]; ok {
				delete(keysToFind, *param.Name)
				log.Printf("✅ Found required parameter: %s", *param.Name)
			}
		}
	}

	// After iterating through all pages, check if any keys are still in keysToFind
	if len(keysToFind) > 0 {
		missingKeys := make([]string, 0, len(keysToFind))
		for k := range keysToFind {
			missingKeys = append(missingKeys, k)
		}
		return fmt.Errorf("parameter validation failed. The following keys are missing: \n - %s", strings.Join(missingKeys, "\n - "))
	}

	log.Println("✅ All required parameters were successfully validated.")
	return nil

}

func loginAWS(ctx context.Context, creds *CredsFile) aws.Config {

	staticProvider := credentials.NewStaticCredentialsProvider(
		creds.CloudProvider.Config.AccessKey,
		creds.CloudProvider.Config.SecretKey,
		creds.CloudProvider.Config.Token,
	)

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(creds.CloudProvider.Config.Region),
		config.WithCredentialsProvider(staticProvider), // <-- This overrides the default chain
	)
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	return cfg
}
