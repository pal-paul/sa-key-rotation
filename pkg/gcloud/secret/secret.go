package secret

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const layout = "2006-01-02T15:04:05Z"

type Secret struct {
	projectID string
	ctx       context.Context
	client    *secretmanager.Client
}
type SecretInterface interface {
	Create(secretName string) error
	AddSecretVersion(secretName string, data []byte) error
	AccessSecretVersion(secretName string, version string) ([]byte, error)
	Exists(secretName string) bool
	DeleteSecretVersion(secretName string, version string) error
}

func New(projectId string) (SecretInterface, error) {
	secret := Secret{}
	secret.projectID = projectId
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return &secret, fmt.Errorf("failed to create secretmanager client: %v", err)
	}

	secret.ctx = ctx
	secret.client = client
	return &secret, nil
}

// Create a new secret with the given name. A secret is a logical
// wrapper around a collection of secret versions. Secret versions hold the
// actual secret material.
func (s *Secret) Create(secretName string) error {
	parent := fmt.Sprintf("projects/%v", s.projectID)

	// Build the request.
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: secretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	// Call the API.
	_, err := s.client.CreateSecret(s.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}
	return nil
}

// AddSecretVersion adds a new secret version to the given secret with the
// provided payload.
func (s *Secret) AddSecretVersion(secretName string, data []byte) error {
	parent := fmt.Sprintf("projects/%v/secrets/%v", s.projectID, secretName)

	// Build the request.
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: parent,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}

	// Call the API.
	_, err := s.client.AddSecretVersion(s.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %v", err)
	}
	return nil
}

// AccessSecretVersion accesses the payload for the given secret version if one
// exists.
func (s *Secret) AccessSecretVersion(secretName string, version string) ([]byte, error) {
	if version == "" {
		version = "latest"
	}
	name := fmt.Sprintf("projects/%v/secrets/%v/versions/%v", s.projectID, secretName, version)

	fmt.Println(name)
	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := s.client.AccessSecretVersion(s.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version: %v", err)
	}
	return result.Payload.Data, nil
}

// Exists Checks if secret exists
func (s *Secret) Exists(secretName string) bool {
	_, err := s.AccessSecretVersion(secretName, "")
	if err != nil {
		return false
	}
	return true
}

// DeleteSecretVersion Deletes specific version of a secret
func (s *Secret) DeleteSecretVersion(secretName string, version string) error {
	destroySecretReq := &secretmanagerpb.DestroySecretVersionRequest{
		Name: fmt.Sprintf("projects/%v/secrets/%v/versions/%v", s.projectID, secretName, version),
	}
	_, err := s.client.DestroySecretVersion(s.ctx, destroySecretReq)
	if err != nil {
		return err
	}
	return nil
}
