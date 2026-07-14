package gcpsecretmanager

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// secretManagerAPI is the subset of the GCP Secret Manager client used by
// the signer. It is defined as an interface so tests can substitute a mock
// implementation without making real GCP API calls or requiring credentials.
//
// The real *secretmanager.Client methods additionally accept variadic
// gax.CallOption arguments; gcpClient adapts that away so callers in this
// package don't need to depend on the gax package directly.
type secretManagerAPI interface {
	AccessSecretVersion(
		ctx context.Context,
		req *secretmanagerpb.AccessSecretVersionRequest,
	) (*secretmanagerpb.AccessSecretVersionResponse, error)

	CreateSecret(
		ctx context.Context,
		req *secretmanagerpb.CreateSecretRequest,
	) (*secretmanagerpb.Secret, error)

	AddSecretVersion(
		ctx context.Context,
		req *secretmanagerpb.AddSecretVersionRequest,
	) (*secretmanagerpb.SecretVersion, error)
}

// gcpClient adapts *secretmanager.Client to the secretManagerAPI interface.
type gcpClient struct {
	c *secretmanager.Client
}

// gcpClient type implements secretManagerAPI.
var _ secretManagerAPI = (*gcpClient)(nil)

func (g *gcpClient) AccessSecretVersion(
	ctx context.Context,
	req *secretmanagerpb.AccessSecretVersionRequest,
) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return g.c.AccessSecretVersion(ctx, req)
}

func (g *gcpClient) CreateSecret(
	ctx context.Context,
	req *secretmanagerpb.CreateSecretRequest,
) (*secretmanagerpb.Secret, error) {
	return g.c.CreateSecret(ctx, req)
}

func (g *gcpClient) AddSecretVersion(
	ctx context.Context,
	req *secretmanagerpb.AddSecretVersionRequest,
) (*secretmanagerpb.SecretVersion, error) {
	return g.c.AddSecretVersion(ctx, req)
}

// newClient builds a real GCP Secret Manager client using Application
// Default Credentials (the GOOGLE_APPLICATION_CREDENTIALS environment
// variable, gcloud user credentials, or the attached service account on
// GCE/GKE/Cloud Run).
func newClient(ctx context.Context) (secretManagerAPI, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create GCP Secret Manager client: %w", err)
	}

	return &gcpClient{c: c}, nil
}
