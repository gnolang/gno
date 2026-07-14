package awssecretsmanager

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// secretsManagerAPI is the subset of the AWS Secrets Manager client used by
// the signer. It is defined as an interface so tests can substitute a mock
// implementation without making real AWS API calls or requiring credentials.
type secretsManagerAPI interface {
	GetSecretValue(
		ctx context.Context,
		params *secretsmanager.GetSecretValueInput,
		optFns ...func(*secretsmanager.Options),
	) (*secretsmanager.GetSecretValueOutput, error)

	CreateSecret(
		ctx context.Context,
		params *secretsmanager.CreateSecretInput,
		optFns ...func(*secretsmanager.Options),
	) (*secretsmanager.CreateSecretOutput, error)
}

// secretsManagerAPI is implemented by *secretsmanager.Client.
var _ secretsManagerAPI = (*secretsmanager.Client)(nil)

// newClient builds a real AWS Secrets Manager client using the standard AWS
// SDK configuration chain (environment variables, shared config/credentials
// files, EC2/ECS instance roles, etc.), optionally pinned to a specific region.
func newClient(ctx context.Context, region string) (secretsManagerAPI, error) {
	var opts []func(*awsconfig.LoadOptions) error
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK configuration: %w", err)
	}

	return secretsmanager.NewFromConfig(awsCfg), nil
}
