package productMapRepository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

type JsonForFrontend struct {
	BucketName string
	ObjectKey  string
}

type MapboxRepo struct {
	logger          ports.Logger
	secrets         secrets
	jsonForFrontend JsonForFrontend
}

func NewMapboxRepo(ctx context.Context, logger ports.Logger, secretsKey string, jsonForFrontend JsonForFrontend) *MapboxRepo {
	return &MapboxRepo{logger: logger, secrets: getSecrets(ctx, logger, secretsKey), jsonForFrontend: jsonForFrontend}
}

func (repo *MapboxRepo) Publish(ctx context.Context, m domain.LocalMap) {
	tempCreds := repo.getMapboxTempCreds(ctx, repo.secrets.MapboxUsername, repo.secrets.MapboxPublicToken)
	repo.uploadToMapboxTempS3(ctx, m.Filepath, tempCreds)
	repo.logger.Debug(ctx, "Uploaded to mapbox's temp s3 succeeded, attempting to notify mapbox.")
	tileset := repo.uploadToMapbox(ctx, tempCreds, repo.secrets.MapboxPublicToken, repo.secrets.MapboxUsername, m.Map.String())
	repo.logger.Info(ctx, "Mapbox was updated with a new tileset, updating frontend.", "tilesetName", tileset)
	repo.updateJSONForFrontend(ctx, m.Map, tileset)
}

func (repo *MapboxRepo) updateJSONForFrontend(ctx context.Context, m domain.Map, tilesetName string) {
	// get existing json file from s3 bucket by key
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		repo.logger.Fatal(ctx, "Could not load aws config", "error", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	getObjectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(repo.jsonForFrontend.BucketName),
		Key:    aws.String(repo.jsonForFrontend.ObjectKey),
	})

	var jsonData []byte
	var mapOptionsList []*conflict_nightlightv1.MapOptions

	if err != nil {
		var nsk *types.NoSuchKey
		// Sort mapOptionsList by date
		if errors.As(err, &nsk) {
			// If the key doesn't exist, create a new file
			mapOptionsList = make([]*conflict_nightlightv1.MapOptions, 0)
		} else {
			repo.logger.Fatal(ctx, "Error when attempting to get json file from s3", "error", err)
		}
	} else {
		defer getObjectOutput.Body.Close()

		jsonData, err = io.ReadAll(getObjectOutput.Body)
		if err != nil {
			repo.logger.Fatal(ctx, "Error when reading s3 object content", "error", err)
		}

		// Unmarshal s3 file into mapOptions
		if err = json.Unmarshal(jsonData, &mapOptionsList); err != nil {
			repo.logger.Fatal(ctx, "Error when unmarshalling s3 object content", "error", err)
		}
	}

	// add a new option to the json file from s3:
	newOption := conflict_nightlightv1.MapOptions{
		DisplayName: fmt.Sprintf("%s %d", time.Month(m.Date.Month), m.Date.Year), // name = m.Date in the format Feb 2022
		Url:         fmt.Sprintf("mapbox://%s", tilesetName),                     // url =  prefix mapbox://   onto the tilsetName
		Key:         m.String(),
	}
	mapOptionsList = updateMapOptionsList(mapOptionsList, &newOption)

	// Marshal updated mapOptionsList back into JSON
	updatedJSON, err := json.Marshal(mapOptionsList)
	if err != nil {
		repo.logger.Fatal(ctx, "Error when marshalling updated map options list", "error", err)
	}

	// Upload the file back to S3
	if _, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(repo.jsonForFrontend.BucketName),
		Key:    aws.String(repo.jsonForFrontend.ObjectKey),
		Body:   bytes.NewReader(updatedJSON),
	}); err != nil {
		repo.logger.Fatal(ctx, "Error when uploading updated json file back to s3", "error", err)
	}
}

func updateMapOptionsList(mapOptionsList []*conflict_nightlightv1.MapOptions,
	newOption *conflict_nightlightv1.MapOptions) []*conflict_nightlightv1.MapOptions {
	// Check if there is an object with the same key, and if there is, replace that object with the new one in mapOptionsList
	found := false
	for i, option := range mapOptionsList {
		if option.Key == newOption.Key {
			mapOptionsList[i] = newOption
			found = true
			break
		}
	}
	if !found {
		mapOptionsList = append(mapOptionsList, newOption)
		// Sort mapOptionsList by date
		sort.Slice(mapOptionsList, func(i, j int) bool {
			dateI, _ := time.Parse("January 2006", mapOptionsList[i].DisplayName)
			dateJ, _ := time.Parse("January 2006", mapOptionsList[j].DisplayName)
			return dateI.Before(dateJ)
		})
	}

	return mapOptionsList
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

type secrets struct {
	MapboxPublicToken string `json:"mapboxPublicToken"`
	MapboxUsername    string `json:"mapboxUsername"`
}

func getSecrets(ctx context.Context, logger ports.Logger, secretsKey string) secrets {
	// TODO i am creating the config in multiple places, where does it make sense to keep functions like this.
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		logger.Fatal(ctx, "Could not load aws config", "error", err)
	}
	client := secretsmanager.NewFromConfig(cfg)
	value, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretsKey,
	})
	if err != nil {
		logger.Fatal(ctx, "Error when attempting to get secret value.", "error", err, "secretsKey", secretsKey)
	}
	var secrets secrets
	err = json.Unmarshal([]byte(*value.SecretString), &secrets)
	if err != nil {
		logger.Fatal(ctx, "Error while attempting to unmarshal the secret string", "error", err)
	}
	return secrets
}

func (repo *MapboxRepo) getMapboxTempCreds(ctx context.Context, username string, token string) mapBoxTempCreds {
	url := fmt.Sprintf("https://api.mapbox.com/uploads/v1/%s/credentials?access_token=%s", username, token)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		repo.logger.Fatal(ctx, "Error when creating request for mapbox Credentials", "error", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		repo.logger.Fatal(ctx, "Error when performing http.Client request for mapbox Credentials", "error", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		repo.logger.Fatal(ctx, "Error when reading the response from mapbox", "error", err)
	}

	var tempCreds mapBoxTempCreds
	if err = json.Unmarshal(body, &tempCreds); err != nil {
		repo.logger.Fatal(ctx, "Error when unmarshalling mapbox credentials", "error", err)
	}
	return tempCreds
}

func (repo *MapboxRepo) uploadToMapboxTempS3(ctx context.Context, localFilepath string, tempAWSCreds mapBoxTempCreds) {
	repo.logger.Info(ctx, "Uploading tif to Mapbox's temp s3 bucket", "localFilepath", localFilepath)
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(tempAWSCreds.AccessKeyId, tempAWSCreds.SecretAccessKey, tempAWSCreds.SessionToken)))
	if err != nil {
		repo.logger.Fatal(ctx, "There was an error when attempting to create a config from the temp mapbox credentials.", "error", err)
	}
	cfg.Region = "us-east-1"

	s3Client := s3.NewFromConfig(cfg)

	file, err := os.Open(localFilepath)
	if err != nil {
		repo.logger.Fatal(ctx, "Error when attempting to open the file to upload to mapbox", "file to upload", localFilepath, "error", err)
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
		repo.logger.Fatal(ctx, "There was an error when attempting to upload the file to mapbox's s3 bucket.", "error", err)
	}
}

func (repo *MapboxRepo) uploadToMapbox(ctx context.Context, tempAWSCreds mapBoxTempCreds, accessToken string, username string, tilesetName string) string {
	url := fmt.Sprintf("https://api.mapbox.com/uploads/v1/%s", username)

	//  tileset: the name passed along to the frontend
	//  name: what is shown in the mapbox ui
	repo.logger.Info(ctx, "Uploading file to mapbox", "tilesetName", tilesetName, "username", username)
	requestBody, err := json.Marshal(map[string]string{
		"url":     fmt.Sprintf("http://%s.s3.amazonaws.com/%s", tempAWSCreds.Bucket, tempAWSCreds.Key),
		"tileset": fmt.Sprintf("%s.%s", username, tilesetName),
		"name":    tilesetName,
	})
	if err != nil {
		repo.logger.Fatal(ctx, "There was an error when formatting the request body.", "error", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		repo.logger.Fatal(ctx, "There was an error when creating the HTTP request", "error", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	query := req.URL.Query()
	query.Add("access_token", accessToken)
	req.URL.RawQuery = query.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		repo.logger.Fatal(ctx, "There was an error when sending the HTTP request", "error", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			repo.logger.Error(ctx, "There was an error when trying to close the response body from the mapbox api", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != 201 {
		repo.logger.Fatal(ctx, "Uploading the tif file via the mapbox api resulted in an unexpected http status code", "statusCode", resp.StatusCode)
	}
	var uploadStatus uploadStatus
	if err = json.NewDecoder(resp.Body).Decode(&uploadStatus); err != nil {
		repo.logger.Fatal(ctx, "Error when trying to unmarshal the response body.", "error", err)
	}
	return uploadStatus.Tileset
}
