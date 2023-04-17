package frontendmapdatarepo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"sort"
	"time"
)

type S3FrontendMapDataRepo struct {
	logger     ports.Logger
	awsClient  awsclient.AWSClient
	bucketName string
	objectKey  string
}

func NewS3FrontendMapDataRepo(ctx context.Context, logger ports.Logger, bucketName string, objectKey string) *S3FrontendMapDataRepo {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		logger.Fatal(ctx, "Error when attempting to load the aws config", "error", err)
	}
	awsClient := awsclient.NewAWSClient(cfg)
	return &S3FrontendMapDataRepo{logger: logger, bucketName: bucketName, objectKey: objectKey, awsClient: awsClient}
}

func (repo *S3FrontendMapDataRepo) List(ctx context.Context) ([]domain.PublishedMap, error) {
	object, err := repo.awsClient.GetFromS3(ctx, repo.bucketName, repo.objectKey)
	if err != nil {
		repo.logger.Error(ctx, "Error when attempting to list objects in s3", "error", err)
		return nil, err
	}
	if object == nil {
		repo.logger.Warn(ctx, "The object was not found in s3", "bucketName", repo.bucketName)
		return nil, nil
	}
	var mapOptionsList []*conflict_nightlightv1.MapOptions
	if err = json.Unmarshal(object, &mapOptionsList); err != nil {
		repo.logger.Error(ctx, "Error when attempting to unmarshal the json", "error", err)
		return nil, err
	}
	publishedMaps := make([]domain.PublishedMap, len(mapOptionsList))
	for i, mapOptions := range mapOptionsList {
		publishedMaps[i] = domain.PublishedMap{
			Map: prototransformers.ProtoToDomain(mapOptions.Map),
			Url: mapOptions.Url,
		}
	}
	return publishedMaps, nil
}

func (repo *S3FrontendMapDataRepo) Delete(ctx context.Context, m domain.Map) error {
	jsonForFrontend, err := repo.awsClient.GetFromS3(ctx, repo.bucketName, repo.objectKey)
	if err != nil {
		return err
	}
	var mapOptionsList []*conflict_nightlightv1.MapOptions
	if err = json.Unmarshal(jsonForFrontend, &mapOptionsList); err != nil {
		repo.logger.Error(ctx, "Error when unmarshalling s3 object content", "error", err)
		return err
	}
	var mapToDelete *conflict_nightlightv1.MapOptions
	for _, mapOptions := range mapOptionsList {
		if mapOptions.Key == m.String() {
			mapToDelete = mapOptions
			break
		}
	}
	if mapToDelete == nil {
		repo.logger.Info(ctx, "The map was not found in the json file", "map", m.String())
		return nil
	}
	mapOptionsList = removeMapOptionFromList(mapOptionsList, mapToDelete)
	jsonForFrontend, err = json.Marshal(mapOptionsList)
	if err != nil {
		repo.logger.Error(ctx, "Error when marshalling the json file", "error", err)
		return err
	}
	if err = repo.awsClient.UploadToS3(ctx, repo.bucketName, repo.objectKey, bytes.NewReader(jsonForFrontend)); err != nil {
		repo.logger.Error(ctx, "Error when attempting to put the json file to s3", "error", err)
		return err
	}
	return nil
}

func removeMapOptionFromList(list []*conflict_nightlightv1.MapOptions, toDelete *conflict_nightlightv1.MapOptions) []*conflict_nightlightv1.MapOptions {
	for i, mapOptions := range list {
		if mapOptions.Key == toDelete.Key {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func (repo *S3FrontendMapDataRepo) Upsert(ctx context.Context, m domain.PublishedMap) error {
	jsonForFrontend, err := repo.awsClient.GetFromS3(ctx, repo.bucketName, repo.objectKey)

	var mapOptionsList []*conflict_nightlightv1.MapOptions

	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			// If the key doesn't exist, create a new file
			mapOptionsList = make([]*conflict_nightlightv1.MapOptions, 0)
		} else {
			repo.logger.Error(ctx, "Error when attempting to get json file from s3", "error", err)
			return err
		}
	} else {
		if err = json.Unmarshal(jsonForFrontend, &mapOptionsList); err != nil {
			repo.logger.Error(ctx, "Error when unmarshalling s3 object content", "error", err)
			return err
		}
	}

	date := time.Date(m.Map.Date.Year, time.Month(m.Map.Date.Month), m.Map.Date.Day, 0, 0, 0, 0, time.UTC)
	protoMap := prototransformers.DomainToProto(m.Map)
	newOption := conflict_nightlightv1.MapOptions{
		DisplayName: fmt.Sprintf("%s %d", date.Format("Jan"), m.Map.Date.Year),
		Url:         m.Url,
		Key:         m.Map.String(),
		Map:         &protoMap,
	}

	mapOptionsList = updateMapOptionsList(mapOptionsList, &newOption)

	updatedJSON, err := json.Marshal(mapOptionsList)
	if err != nil {
		repo.logger.Error(ctx, "Error when marshalling updated map options list", "error", err)
		return err
	}

	err = repo.awsClient.UploadToS3(ctx, repo.bucketName, repo.objectKey, bytes.NewReader(updatedJSON))
	if err != nil {
		repo.logger.Error(ctx, "Error when uploading updated json file back to s3", "error", err)
		return err
	}
	return nil
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
	}
	// Sort mapOptionsList by date
	sort.Slice(mapOptionsList, func(i, j int) bool {
		dateI, _ := time.Parse("Jan 2006", mapOptionsList[i].DisplayName)
		dateJ, _ := time.Parse("Jan 2006", mapOptionsList[j].DisplayName)
		return dateI.Before(dateJ)
	})

	return mapOptionsList
}
