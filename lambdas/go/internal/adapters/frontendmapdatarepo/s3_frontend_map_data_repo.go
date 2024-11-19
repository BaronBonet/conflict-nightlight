package frontendmapdatarepo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3FrontendMapDataRepo struct {
	logger     ports.Logger
	awsClient  awsclient.AWSClient
	bucketName string
	objectKey  string
}

func NewS3FrontendMapDataRepo(
	ctx context.Context,
	logger ports.Logger,
	bucketName string,
	objectKey string,
) ports.FrontendMapDataRepo {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		logger.Fatal(ctx, "Error when attempting to load the aws config", "error", err)
	}
	awsClient := awsclient.NewAWSClient(cfg)
	return &s3FrontendMapDataRepo{logger: logger, bucketName: bucketName, objectKey: objectKey, awsClient: awsClient}
}

func (repo *s3FrontendMapDataRepo) List(ctx context.Context) ([]domain.PublishedMap, error) {
	object, err := repo.awsClient.GetFromS3(ctx, repo.bucketName, repo.objectKey)
	if err != nil {
		repo.logger.Error(ctx, "Error when attempting to list objects in s3", "error", err)
		return nil, err
	}
	if object == nil {
		repo.logger.Warn(ctx, "The object was not found in s3", "bucketName", repo.bucketName)
		return nil, nil
	}
	var boundedMapOptionsList []*conflict_nightlightv1.BoundedMapOptions
	if err = json.Unmarshal(object, &boundedMapOptionsList); err != nil {
		repo.logger.Error(ctx, "Error when attempting to unmarshal the json", "error", err)
		return nil, err
	}
	// Calculate total length of all maps across all bounds
	totalLength := 0
	for _, boundedMaps := range boundedMapOptionsList {
		totalLength += len(boundedMaps.GetMapsOptions())
	}

	publishedMaps := make([]domain.PublishedMap, 0, totalLength)
	for _, boundedMaps := range boundedMapOptionsList {
		repo.logger.Debug(ctx, "Found bounded map", "bounds", boundedMaps.GetBounds())
		for _, mapOptions := range boundedMaps.GetMapsOptions() {
			publishedMaps = append(publishedMaps, domain.PublishedMap{
				Map: prototransformers.ProtoToDomain(mapOptions.GetMap()),
				Url: mapOptions.GetUrl(),
			})
		}
	}
	return publishedMaps, nil
}
func (repo *s3FrontendMapDataRepo) Delete(ctx context.Context, m domain.Map) error {
	jsonForFrontend, err := repo.awsClient.GetFromS3(ctx, repo.bucketName, repo.objectKey)
	if err != nil {
		return err
	}
	var boundedMapOptionsList []*conflict_nightlightv1.BoundedMapOptions
	if err = json.Unmarshal(jsonForFrontend, &boundedMapOptionsList); err != nil {
		repo.logger.Error(ctx, "Error when unmarshalling s3 object content", "error", err)
		return err
	}

	mapKey := m.String()
	found := false
	// Iterate through each bounded map options
	for i, boundedMaps := range boundedMapOptionsList {
		mapOptions := boundedMaps.GetMapsOptions()
		for j, option := range mapOptions {
			if option.GetKey() == mapKey {
				// Remove the map option from the slice
				boundedMaps.MapsOptions = append(mapOptions[:j], mapOptions[j+1:]...)
				found = true
				// If this was the last map option in this bounds, remove the entire bounded map option
				if len(boundedMaps.GetMapsOptions()) == 0 {
					boundedMapOptionsList = append(boundedMapOptionsList[:i], boundedMapOptionsList[i+1:]...)
				}
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		repo.logger.Info(ctx, "The map was not found in the json file", "map", mapKey)
		return nil
	}

	jsonForFrontend, err = json.Marshal(boundedMapOptionsList)
	if err != nil {
		repo.logger.Error(ctx, "Error when marshalling the json file", "error", err)
		return err
	}

	if err = repo.awsClient.UploadToS3(ctx,
		repo.bucketName,
		repo.objectKey,
		bytes.NewReader(jsonForFrontend)); err != nil {
		repo.logger.Error(ctx, "Error when attempting to put the json file to s3", "error", err)
		return err
	}
	return nil
}

func (repo *s3FrontendMapDataRepo) Upsert(ctx context.Context, m domain.PublishedMap) error {
	jsonForFrontend, err := repo.awsClient.GetFromS3(ctx, repo.bucketName, repo.objectKey)

	var boundedMapOptionsList []*conflict_nightlightv1.BoundedMapOptions

	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			// If the key doesn't exist, create an empty array that will become the new file
			boundedMapOptionsList = make([]*conflict_nightlightv1.BoundedMapOptions, 0)
		} else {
			repo.logger.Error(ctx, "Error when attempting to get json file from s3", "error", err)
			return err
		}
	} else {
		if err = json.Unmarshal(jsonForFrontend, &boundedMapOptionsList); err != nil {
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

	boundedMapOptionsList = updateBoundedMapOptions(boundedMapOptionsList, &newOption)

	updatedJSON, err := json.Marshal(boundedMapOptionsList)
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

func updateBoundedMapOptions(
	boundedMapOptions []*conflict_nightlightv1.BoundedMapOptions, newOption *conflict_nightlightv1.MapOptions,
) []*conflict_nightlightv1.BoundedMapOptions {
	found := false
	for i, boundedMapOption := range boundedMapOptions {
		if boundedMapOption.GetBounds() == newOption.GetMap().GetBounds() {
			mapOptions := updateMapOptionsList(boundedMapOption.GetMapsOptions(), newOption)
			found = true
			boundedMapOptions[i].MapsOptions = mapOptions
			break
		}
	}
	if !found {
		boundedMapOptions = append(boundedMapOptions, &conflict_nightlightv1.BoundedMapOptions{
			Bounds:      newOption.GetMap().GetBounds(),
			MapsOptions: []*conflict_nightlightv1.MapOptions{newOption},
		})
	}
	return boundedMapOptions
}

func updateMapOptionsList(mapOptionsList []*conflict_nightlightv1.MapOptions,
	newOption *conflict_nightlightv1.MapOptions) []*conflict_nightlightv1.MapOptions {
	// Check if there is an object with the same key, and if there is,
	// replace that object with the new one in mapOptionsList
	found := false
	for i, option := range mapOptionsList {
		if option.GetKey() == newOption.GetKey() {
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
		dateI, _ := time.Parse("Jan 2006", mapOptionsList[i].GetDisplayName())
		dateJ, _ := time.Parse("Jan 2006", mapOptionsList[j].GetDisplayName())
		return dateI.Before(dateJ)
	})

	return mapOptionsList
}
