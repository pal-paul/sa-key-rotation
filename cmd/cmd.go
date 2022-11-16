package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pal-paul/sa-key-rotation/pkg/gcloud/serviceaccount"
	"github.com/pal-paul/sa-key-rotation/pkg/github"
)

var (
	owner string
	repo  string
	token string
)

func main() {
	var (
		sa          string
		secret_name string
	)
	owner = os.Getenv("INPUT_OWNER")
	repo = os.Getenv("INPUT_REPO")
	token = os.Getenv("INPUT_TOKEN")

	sa = os.Getenv("INPUT_SERVICE_ACCOUNT_EMAIL")
	secret_name = os.Getenv("INPUT_SECRET_NAME")

	client, err := serviceaccount.New(sa)
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
		log.Fatal("no keys found for service account to rotate")
		return
	}
	if len(keys) > 1 {
		log.Fatal("more than one key found for service account, not allowed to rotate")
		return
	}
	for _, key := range keys {
		cred, err := client.CreateKey()
		if err != nil {
			log.Fatal("failed to create key: ", err)
			return
		}
		gitClient, err := github.New(owner, repo, token)
		if err != nil {
			log.Fatal("failed to create github client: ", err)
			return
		}
		secretData, err := json.Marshal(cred)
		if err != nil {
			log.Fatal("failed to marshal secret data: ", err)
			return
		}
		err = gitClient.CreateUpdateRepoSecret(secret_name, string(secretData))
		if err != nil {
			log.Fatal("failed to create or update repo secret: ", err)
			return
		}
		err = client.DeleteKey(key.Name)
		if err != nil {
			log.Fatal("failed to delete key: ", err)
			return
		}
	}
}
