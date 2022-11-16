package github

import (
	"time"
)

type RepoSecrets struct {
	TotalCount int      `json:"total_count"`
	Secrets    []Secret `json:"secrets"`
}

type Secret struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// key represents a secret of key and id.
type key struct {
	Key   string `json:"key"`
	KeyID string `json:"key_id"`
}

// encryptedSecret represents a secret that is encrypted using a public key.
type encryptedSecret struct {
	KeyID          string `json:"key_id"`
	EncryptedValue string `json:"encrypted_value"`
}
