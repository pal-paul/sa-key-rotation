package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	sodium "github.com/GoKillers/libsodium-go/cryptobox"
)

const baseUrl = "https://api.github.com"
const basePath = "api"

type github struct {
	token string
	owner string
	repo  string
	key   key
}
type GithubInterface interface {
	CreateUpdateRepoSecret(SecretName string, secretValue string) error
	RepoSecrets() (RepoSecrets, error)
}

func New(owner string, repo string, token string) (GithubInterface, error) {
	git := github{
		token: token,
		owner: owner,
		repo:  repo,
	}
	err := git.getPublicKey()
	if err != nil {
		return &git, err
	}
	return &git, nil
}

// Creates or updates a repository secret with value.
func (g *github) CreateUpdateRepoSecret(SecretName string, secretValue string) error {
	path := "actions/secrets/" + SecretName
	encryptSecretData, err := g.getEncryptSecretData(secretValue)
	if err != nil {
		return err
	}
	reqBody, err := json.Marshal(encryptSecretData)
	if err != nil {
		return err
	}
	_, err = g.put(path, reqBody, nil)
	if err != nil {
		return err
	}
	return nil
}

// Lists all secrets available in a repository without revealing their encrypted values
func (g *github) RepoSecrets() (RepoSecrets, error) {
	var rs RepoSecrets
	path := "actions/secrets"
	body, err := g.get(path, nil)
	if err != nil {
		return rs, err
	}
	err = json.Unmarshal(body, &rs)
	if err != nil {
		return rs, err
	}
	return rs, nil
}

// Gets public key, which need to encrypt a secret before can create or update secrets.
func (g *github) getPublicKey() error {
	path := "actions/secrets/public-key"
	var gKey key
	body, err := g.get(path, nil)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &gKey)
	if err != nil {
		return err
	}
	g.key = gKey
	return nil
}

func (g *github) getEncryptSecretData(secretValue string) (*encryptedSecret, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(g.key.Key)
	if err != nil {
		return nil, fmt.Errorf("base64.StdEncoding.DecodeString was unable to decode public key: %v", err)
	}
	secretBytes := []byte(secretValue)
	encryptedBytes, exit := sodium.CryptoBoxSeal(secretBytes, decodedPublicKey)
	if exit != 0 {
		return nil, errors.New("sodium.CryptoBoxSeal exited with non zero exit code")
	}
	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
	encryptedSecret := &encryptedSecret{
		KeyID:          g.key.KeyID,
		EncryptedValue: encryptedString,
	}
	return encryptedSecret, nil
}

func (g *github) get(path string, qs url.Values) ([]byte, error) {
	gUrl := baseUrl + "/repos/" + g.owner + "/" + g.repo + "/" + path

	u, err := url.Parse(gUrl)
	if qs != nil {
		u.RawQuery = qs.Encode()
	}

	gUrl = u.String()

	// Create a Bearer string by appending string access token
	var token = "token " + g.token
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, gUrl, nil)
	if err != nil {
		return nil, err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read put resp from %s: %v\n", gUrl, err)
		return nil, err
	}
	if resp.StatusCode == 200 {
		return body, nil
	} else {
		return nil, fmt.Errorf("failed action on github status  %v", resp.StatusCode)
	}
}

func (g *github) post(path string, reqBody []byte, qs url.Values) ([]byte, error) {
	gUrl := baseUrl + "/repos/" + g.owner + "/" + g.repo + "/" + path

	u, err := url.Parse(gUrl)
	if qs != nil {
		u.RawQuery = qs.Encode()
	}

	gUrl = u.String()

	// Create a Bearer string by appending string access token
	var token = "token " + g.token
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, gUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to post requst from %s: %v\n", gUrl, err)
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 210 {
		return body, nil
	} else {
		return nil, fmt.Errorf("failed action on github status  %v", resp.StatusCode)
	}
}

func (g *github) put(path string, reqBody []byte, qs url.Values) ([]byte, error) {
	gUrl := baseUrl + "/repos/" + g.owner + "/" + g.repo + "/" + path

	u, err := url.Parse(gUrl)
	if qs != nil {
		u.RawQuery = qs.Encode()
	}

	gUrl = u.String()

	// Create a Bearer string by appending string access token
	var token = "token " + g.token
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPut, gUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 201 || resp.StatusCode == 204 {
		return body, nil
	} else {
		return nil, fmt.Errorf("failed action on github status  %v", resp.StatusCode)
	}
}
