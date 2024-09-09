package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SecretManager struct {
	client *secretmanager.Client
}

func NewSecretManager() *SecretManager {
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create secret manager client: %v", err)
	}

	return &SecretManager{
		client: client,
	}
}

func (sm *SecretManager) GetSecret(secretName string) map[string]string {
	ctx := context.Background()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(
			"projects/sityex-dev/secrets/%s/versions/latest",
			secretName,
		),
	}

	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("Failed to access secret version: %v", err)
	}

	secretPayload := result.Payload.Data
	secretMap := map[string]string{}
	err = json.Unmarshal(secretPayload, &secretMap)
	if err != nil {
		log.Fatalf("Failed to unmarshal secret payload: %v", err)
	}

	return secretMap
}
