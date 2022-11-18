package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pal-paul/sa-key-rotation/pkg/gcloud/secret"
	"github.com/pal-paul/sa-key-rotation/pkg/gcloud/serviceaccount"
	"github.com/pal-paul/sa-key-rotation/pkg/github"
)

var (
	owner string
	repo  string
	token string

	projectId              string
	sa                     string
	github_secret_key_name string
	gcp_secret_key_name    string
)

func main() {
	owner = os.Getenv("INPUT_OWNER")
	repo = os.Getenv("INPUT_REPO")
	token = os.Getenv("INPUT_TOKEN")
	projectId = os.Getenv("INPUT_PROJECT_ID")

	sa = os.Getenv("INPUT_SERVICE_ACCOUNT_NAME")
	github_secret_key_name = os.Getenv("INPUT_GITHUB_SECRET_KEY_NAME")
	gcp_secret_key_name = os.Getenv("INPUT_GCP_SECRET_KEY_NAME")

	client, err := serviceaccount.New(sa + "@" + projectId + ".iam.gserviceaccount.com")
	if err != nil {
		log.Fatal("failed to create client: ", err)
		return
	}
	keys, err := client.Keys()
	if err != nil {
		log.Fatal("failed to get service account email keys")
		return
	}
	if len(keys) == 0 {
		err = CreateRotate(client)
		if err != nil {
			log.Fatal("failed to rotate service account key: ", err)
			return
		}
		return
	}
	if len(keys) > 1 {
		log.Fatal("more than one key found for service account, not allowed to rotate")
		return
	}
	for _, key := range keys {
		err = CreateRotate(client)
		if err != nil {
			log.Fatal("failed to rotate service account key: ", err)
			return
		}
		err = client.DeleteKey(key.FullName)
		if err != nil {
			log.Fatal("failed to delete key: ", err)
			return
		}
	}
}

func CreateRotate(client serviceaccount.ServiceAccountInterface) error {
	cred, err := client.CreateKey()
	if err != nil {
		return err
	}
	fmt.Println("New key created")
	gitClient, err := github.New(owner, repo, token)
	if err != nil {
		return err
	}
	secretData, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	err = gitClient.CreateUpdateRepoSecret(github_secret_key_name, string(secretData))
	if err != nil {
		return err
	}
	fmt.Println("New key added to github repo")
	if gcp_secret_key_name != "" {
		secretManager, err := secret.New(projectId)
		if err != nil {
			return err
		}
		if !secretManager.Exists(gcp_secret_key_name) {
			secretManager.Create(gcp_secret_key_name)
			fmt.Println("New secret created in gcp secret manager")
		}
		secretManager.AddSecretVersion(gcp_secret_key_name, secretData)
		fmt.Println("New key added to gcp secret manager")
	}
	return nil
}
