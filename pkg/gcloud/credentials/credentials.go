package credentials

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type Credentials struct{}

func (c *Credentials) GetProjectId() (string, error) {
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", err
	}

	return credentials.ProjectID, nil
}
