package serviceaccount

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	iam "google.golang.org/api/iam/v1"
	"net/url"
	"path"
	"time"
)

const layout = "2006-01-02T15:04:05Z"

type ServiceAccount struct {
	serviceAccountEmail string
	ctx                 context.Context
	service             *iam.Service
}
type ServiceAccountInterface interface {
	CreateKey() (Credential, error)
	Keys() ([]Key, error)
	DeleteKey(fullKeyName string) error
}

func New(serviceAccountEmail string) (ServiceAccountInterface, error) {
	sa := ServiceAccount{
		serviceAccountEmail: serviceAccountEmail,
	}
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		return &sa, fmt.Errorf("iam.NewService: %v", err)
	}
	sa.ctx = ctx
	sa.service = service
	return &sa, nil
}

// createKey creates a service account key.
func (sa *ServiceAccount) CreateKey() (Credential, error) {
	var cred Credential
	resource := "projects/-/serviceAccounts/" + sa.serviceAccountEmail
	request := &iam.CreateServiceAccountKeyRequest{}
	key, err := sa.service.Projects.ServiceAccounts.Keys.Create(resource, request).Do()
	if err != nil {
		return cred, fmt.Errorf("Projects.ServiceAccounts.Keys.Create: %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(key.PrivateKeyData)
	if err != nil {
		fmt.Println("decode error:", err)
		return cred, fmt.Errorf("failed to decoded key: %v", err)
	}
	err = json.Unmarshal(decoded, &cred)
	if err != nil {
		return cred, err
	}

	return cred, nil
}

// listKey lists a service account's keys.
func (sa *ServiceAccount) Keys() ([]Key, error) {
	resource := "projects/-/serviceAccounts/" + sa.serviceAccountEmail
	response, err := sa.service.Projects.ServiceAccounts.Keys.List(resource).Do()
	if err != nil {
		return nil, fmt.Errorf("Projects.ServiceAccounts.Keys.List: %v", err)
	}

	var keys []Key
	for _, key := range response.Keys {
		if key.KeyType == "USER_MANAGED" {
			createdDate, err := time.Parse(layout, key.ValidAfterTime)
			if err != nil {
				return nil, fmt.Errorf("ValidAfterTime invalid: %v", err)
			}
			keyNameParse, err := url.Parse(key.Name)
			if err != nil {
				return nil, fmt.Errorf("Key name invalid: %v", err)
			}

			keys = append(keys, Key{
				Name:      path.Base(keyNameParse.Path),
				FullName:  key.Name,
				KeyType:   key.KeyType,
				CreatedOn: createdDate,
			})
		}
	}
	return keys, nil
}

// DeleteKey deletes a service account key.
func (sa *ServiceAccount) DeleteKey(fullKeyName string) error {
	_, err := sa.service.Projects.ServiceAccounts.Keys.Delete(fullKeyName).Do()
	if err != nil {
		return fmt.Errorf("Projects.ServiceAccounts.Keys.Delete: %v", err)
	}
	return nil
}
