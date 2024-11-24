package maptileserverrepo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mitchellh/mapstructure"
)

type mapBoxTileServerRepo struct {
	logger    ports.Logger
	awsClient awsclient.AWSClient
	secrets   mapboxSecrets
}

type mapboxSecrets struct {
	MapboxPublicToken string `json:"mapboxPublicToken"`
	MapboxUsername    string `json:"mapboxUsername"`
}

func NewMapboxTileServerRepo(ctx context.Context, logger ports.Logger, secretsKey string) ports.MapTileServerRepo {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		logger.Fatal(ctx, "Error when attempting to load the aws config", "error", err)
	}
	awsClient := awsclient.NewAWSClient(cfg)
	rawSecrets, err := awsClient.GetSecretFromSecretsManager(ctx, secretsKey)
	if err != nil {
		logger.Fatal(ctx, "Error when getting secrets from secrets manager", "secretKey", secretsKey, "error", err)
	}
	var secrets mapboxSecrets
	err = mapstructure.Decode(rawSecrets, &secrets)
	if err != nil {
		logger.Fatal(ctx, "Error when getting secrets from secrets manager", "secretKey", secretsKey, "error", err)
	}
	if secrets.MapboxPublicToken == "" || secrets.MapboxUsername == "" {
		logger.Fatal(ctx, "Secrets were not filled", "secretKey", secretsKey)
	}
	return &mapBoxTileServerRepo{logger: logger, secrets: secrets, awsClient: awsClient}
}

func (repo *mapBoxTileServerRepo) Publish(ctx context.Context, m domain.LocalMap) (*domain.PublishedMap, error) {
	tempCreds, err := repo.getMapboxTempCreds(ctx, repo.secrets.MapboxUsername, repo.secrets.MapboxPublicToken)
	if err != nil {
		repo.logger.Error(ctx, "Error when getting temp creds from mapbox", "error", err)
		return nil, err
	}
	if tempCreds == nil {
		repo.logger.Error(ctx, "Error when getting temp creds from mapbox", "error", errors.New("temp creds were nil"))
		return nil, errors.New("temp creds were nil")
	}
	repo.uploadToMapboxTempS3(ctx, m.Filepath, *tempCreds)
	repo.logger.Debug(ctx, "Uploaded to mapbox's temp s3 succeeded, attempting to notify mapbox.")
	tileset, err := repo.uploadToMapbox(
		ctx,
		*tempCreds,
		repo.secrets.MapboxPublicToken,
		repo.secrets.MapboxUsername,
		m.Map.String(),
	)
	if err != nil {
		repo.logger.Error(ctx, "Error when uploading to mapbox", "error", err)
		return nil, err
	}
	repo.logger.Info(ctx, "Mapbox was updated with a new tileset, updating frontend.", "tilesetName", tileset)
	return &domain.PublishedMap{Map: m.Map, Url: fmt.Sprintf("mapbox://%s", tileset)}, nil
}

func (repo *mapBoxTileServerRepo) Delete(ctx context.Context, m domain.Map) error {
	repo.logger.Info(ctx, "Deleting map from mapbox", "map", m.String())
	tilesetID := fmt.Sprintf("%s.%s", repo.secrets.MapboxUsername, m.String())

	url := fmt.Sprintf("https://api.mapbox.com/tilesets/v1/%s", tilesetID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		repo.logger.Error(ctx, "Error when creating request for deleting Mapbox tileset", "error", err)
		return err
	}

	query := req.URL.Query()
	query.Add("access_token", repo.secrets.MapboxPublicToken)
	req.URL.RawQuery = query.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		repo.logger.Error(ctx, "Error when performing http.Client request for deleting Mapbox tileset", "error", err)
		return err
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			repo.logger.Error(
				ctx,
				"There was an error when trying to close the response body from the mapbox api",
				"error",
				err,
			)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		repo.logger.Error(
			ctx,
			"Deleting the tileset via the Mapbox API resulted in an unexpected http status code",
			"statusCode",
			resp.StatusCode,
			"responseBody",
			resp.Body,
		)
		return errors.New("unexpected http status code was returned from Mapbox API")
	}
	return nil
}

type mapBoxTempCreds struct {
	AccessKeyId     string `json:"accessKeyId"`
	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	URL             string `json:"url"`
}

type uploadStatus struct {
	Complete bool      `json:"complete"`
	Tileset  string    `json:"tileset"`
	Error    *string   `json:"error"`
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Modified time.Time `json:"modified"`
	Created  time.Time `json:"created"`
	Owner    string    `json:"owner"`
	Progress int       `json:"progress"`
}

func (repo *mapBoxTileServerRepo) getMapboxTempCreds(
	ctx context.Context,
	username string,
	token string,
) (*mapBoxTempCreds, error) {
	url := fmt.Sprintf("https://api.mapbox.com/uploads/v1/%s/credentials?access_token=%s", username, token)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		repo.logger.Error(ctx, "Error when creating request for mapbox Credentials", "error", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		repo.logger.Error(ctx, "Error when performing http.Client request for mapbox Credentials", "error", err)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		repo.logger.Error(ctx, "Error when reading the response from mapbox", "error", err)
		return nil, err
	}

	var tempCreds mapBoxTempCreds
	if err = json.Unmarshal(body, &tempCreds); err != nil {
		repo.logger.Error(ctx, "Error when unmarshalling mapbox credentials", "error", err)
		return nil, err
	}

	if tempCreds.AccessKeyId == "" {
		repo.logger.Error(ctx, "There was an issue when extracting the temporary credentials", "response", string(body))
		return nil, errors.New("there was an issue when extracting the temporary credentials from mapbox")
	}
	return &tempCreds, nil
}

func (repo *mapBoxTileServerRepo) uploadToMapboxTempS3(
	ctx context.Context,
	localFilepath string,
	tempAWSCreds mapBoxTempCreds,
) {
	repo.logger.Debug(ctx, "Uploading tif to Mapbox's temp s3 bucket", "localFilepath", localFilepath)
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				tempAWSCreds.AccessKeyId,
				tempAWSCreds.SecretAccessKey,
				tempAWSCreds.SessionToken,
			),
		),
	)
	if err != nil {
		repo.logger.Fatal(
			ctx,
			"There was an error when attempting to create a config from the temp mapbox credentials.",
			"error",
			err,
		)
	}
	cfg.Region = "us-east-1"

	s3Client := s3.NewFromConfig(cfg)

	file, err := os.Open(localFilepath)
	if err != nil {
		repo.logger.Fatal(
			ctx,
			"Error when attempting to open the file to upload to mapbox",
			"file to upload",
			localFilepath,
			"error",
			err,
		)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			repo.logger.Error(ctx, "Error when trying to close the file we uploaded")
		}
	}(file)

	if _, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(tempAWSCreds.Bucket),
		Key:    aws.String(tempAWSCreds.Key),
		Body:   file,
	}); err != nil {
		repo.logger.Fatal(
			ctx,
			"There was an error when attempting to upload the file to mapbox's s3 bucket.",
			"error",
			err,
		)
	}
}

func (repo *mapBoxTileServerRepo) uploadToMapbox(
	ctx context.Context,
	tempAWSCreds mapBoxTempCreds,
	accessToken string,
	username string,
	tilesetName string,
) (string, error) {
	url := fmt.Sprintf("https://api.mapbox.com/uploads/v1/%s", username)

	//  tileset: the name passed along to the frontend
	//  name: what is shown in the mapbox ui
	repo.logger.Debug(ctx, "Uploading file to mapbox", "tilesetName", tilesetName, "username", username)
	requestBody, err := json.Marshal(map[string]string{
		"url":     fmt.Sprintf("http://%s.s3.amazonaws.com/%s", tempAWSCreds.Bucket, tempAWSCreds.Key),
		"tileset": fmt.Sprintf("%s.%s", username, tilesetName),
		"name":    tilesetName,
	})
	if err != nil {
		repo.logger.Error(ctx, "when formatting the request body.", "error", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		repo.logger.Error(ctx, "when creating the HTTP request", "error", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	query := req.URL.Query()
	query.Add("access_token", accessToken)
	req.URL.RawQuery = query.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		repo.logger.Error(ctx, "when sending the HTTP request", "error", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			repo.logger.Error(
				ctx,
				"There was an error when trying to close the response body from the mapbox api",
				"error",
				err,
			)
		}
	}(resp.Body)

	if resp.StatusCode != 201 {
		repo.logger.Error(
			ctx,
			"Uploading the tif file via the mapbox api resulted in an unexpected http status code",
			"statusCode",
			resp.StatusCode,
		)
		return "", errors.New("unexpected http status code was returned from mapbox api")
	}
	var uploadStatus uploadStatus
	if err = json.NewDecoder(resp.Body).Decode(&uploadStatus); err != nil {
		repo.logger.Error(ctx, "when trying to unmarshal the response body.", "error", err)
		return "", err
	}
	return uploadStatus.Tileset, nil
}
