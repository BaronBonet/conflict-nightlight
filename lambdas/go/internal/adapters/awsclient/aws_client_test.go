package awsclient

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSecretFromSecretsManager(t *testing.T) {
	ctx := context.Background()
	secretKey := "test-secret-key"

	type mapboxSecrets struct {
		MapboxPublicToken string `json:"mapboxPublicToken"`
		MapboxUsername    string `json:"mapboxUsername"`
	}

	expectedSecrets := mapboxSecrets{
		MapboxPublicToken: "test-public-token",
		MapboxUsername:    "test-username",
	}

	secretString, _ := json.Marshal(expectedSecrets)
	secretStringAsStr := string(secretString)

	mockSecretsManagerClient := NewMockSecretsManagerClientInterface(t)
	mockSecretsManagerClient.On("GetSecretValue", ctx, mock.AnythingOfType("*secretsmanager.GetSecretValueInput")).
		Return(&secretsmanager.GetSecretValueOutput{
			SecretString: &secretStringAsStr,
		}, nil)

	awsClient := &awsClient{
		secretsManagerClient: mockSecretsManagerClient,
	}

	rawSecrets, err := awsClient.GetSecretFromSecretsManager(ctx, secretKey)
	assert.NoError(t, err)

	var secrets mapboxSecrets
	err = mapstructure.Decode(rawSecrets, &secrets)
	assert.NoError(t, err)

	assert.Equal(t, expectedSecrets, secrets)
	mockSecretsManagerClient.AssertExpectations(t)
}
