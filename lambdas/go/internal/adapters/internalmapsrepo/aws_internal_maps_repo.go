package internalmapsrepo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	awsclient "github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/google/uuid"
)

type AWSMapsRepo struct {
	bucket           Bucket
	logger           ports.Logger
	awsClient        awsclient.AWSClient
	messageQueueName string
	tmpWriteDir      string
}

type Bucket struct {
	bucketName  string
	metadataKey string
}

func NewAWSInternalMapsRepository(
	logger ports.Logger,
	repoBucketName string,
	metadataKey string,
	persistRequestQueueName string,
	tmpWriteDir string,
	awsClient awsclient.AWSClient,
) ports.InternalMapRepo {
	return &AWSMapsRepo{logger: logger, bucket: Bucket{
		bucketName:  repoBucketName,
		metadataKey: metadataKey,
	}, messageQueueName: persistRequestQueueName,
		tmpWriteDir: tmpWriteDir,
		awsClient:   awsClient}
}

func (repo *AWSMapsRepo) List(
	ctx context.Context,
	desiredProvider domain.MapProvider,
	desiredBounds domain.Bounds,
	desiredMapType domain.MapType,
) ([]domain.Map, error) {

	objects, err := repo.awsClient.ListObjectsInS3(ctx, repo.bucket.bucketName)
	if err != nil {
		return nil, err
	}
	if objects == nil {
		repo.logger.Warn(ctx, "There were no objects found in the s3 bucket", "bucketName", repo.bucket.bucketName)
		return nil, nil
	}

	var existingMaps []domain.Map
	var wg sync.WaitGroup
	mapChan := make(chan *domain.Map)

	for _, objectKey := range objects {
		wg.Add(1)
		go func(objKey string) {
			defer wg.Done()
			processObjectKey(ctx, repo, objKey, desiredProvider, desiredBounds, desiredMapType, mapChan)
		}(objectKey)
	}

	go func() {
		wg.Wait()
		close(mapChan)
	}()

	for resultMap := range mapChan {
		existingMaps = append(existingMaps, *resultMap)
	}
	return sortMapsByDate(existingMaps), nil
}

func (repo *AWSMapsRepo) Download(ctx context.Context, m domain.Map) (*domain.LocalMap, error) {
	key := createKeyFromMap(m)
	localFilepath := fmt.Sprintf("%s/%s.tif", repo.tmpWriteDir, uuid.NewString())
	data, err := repo.awsClient.GetFromS3(ctx, repo.bucket.bucketName, key)

	if err != nil {
		repo.logger.Error(ctx, "Couldn't download file", "key", "bucket", repo.bucket.bucketName, key, "error", err)
		return nil, err
	}
	file, err := os.Create(localFilepath)
	if err != nil {
		repo.logger.Error(ctx, "Couldn't create file", "localFilepath", localFilepath, "error", err)
		return nil, err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		repo.logger.Error(ctx, "Couldn't write contents from s3", "key", key, "error", err)
		return nil, err
	}
	return &domain.LocalMap{
		Filepath: localFilepath,
		Map:      m,
	}, nil
}

// Create puts a message on a sqs queue which will be picked up by a python application responsible for downloading the tif file
//
//	TODO this needs to be implemented also for the processing step
func (repo *AWSMapsRepo) Create(ctx context.Context, m domain.Map) error {
	protoMap := prototransformers.DomainToProto(m)
	message := conflict_nightlightv1.RequestWrapper_DownloadAndCropRawTifRequest{
		DownloadAndCropRawTifRequest: &conflict_nightlightv1.DownloadAndCropRawTifRequest{Map: &protoMap},
	}
	err := repo.awsClient.PublishMessageToSQS(ctx, repo.messageQueueName, &message)
	return err
}

func (repo *AWSMapsRepo) Delete(ctx context.Context, m domain.Map) error {
	key := createKeyFromMap(m)
	err := repo.awsClient.DeleteFromS3(ctx, repo.bucket.bucketName, key)
	if err != nil {
		repo.logger.Error(ctx, "Couldn't delete file", "key", key, "error", err)
		return err
	}
	return nil
}

func extractDateFromKey(filePath string) (*domain.Date, error) {
	fileName := filepath.Base(filePath)
	parts := strings.Split(fileName, ".")[0]
	dateParts := strings.Split(parts, "_")

	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return nil, errors.New("failed to parse year. error: " + err.Error())
	}

	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		return nil, errors.New("failed to parse month. error: " + err.Error())
	}

	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		return nil, errors.New("failed to parse day. error: " + err.Error())
	}
	return &domain.Date{
		Day:   day,
		Month: time.Month(month),
		Year:  year,
	}, nil
}

func processObjectKey(
	ctx context.Context,
	repo *AWSMapsRepo,
	objKey string,
	desiredProvider domain.MapProvider,
	desiredBounds domain.Bounds,
	desiredMapType domain.MapType,
	mapChan chan<- *domain.Map,
) {
	directory := filepath.Dir(objKey)
	extension := filepath.Ext(objKey)
	if directory == "" || extension != ".tif" {
		repo.logger.Warn(ctx, "The objectKey was not in the expected format", "objectKey", objKey)
		return
	}
	provider, bounds, mapType, err := convertStringPathToDomain(directory)
	if err != nil {
		repo.logger.Error(
			ctx,
			"There was an error when attempting to convert the objectKey to a domain object",
			"error",
			err.Error(),
			"objectKey",
			objKey,
		)
		return
	}
	if isDesiredObject(*provider, desiredProvider, *bounds, desiredBounds, *mapType, desiredMapType) {
		date, err := extractDateFromKey(objKey)
		if err != nil {
			repo.logger.Error(
				ctx,
				"There was an error when attempting to extract the date from the objectKey",
				"error",
				err.Error(),
				"objectKey",
				objKey,
			)
			return
		}

		url, err := repo.awsClient.GetObjectMetadataInS3(ctx, repo.bucket.bucketName, objKey, repo.bucket.metadataKey)

		if err != nil {
			repo.logger.Error(
				ctx,
				"There was an error when attempting to get the metadata from the object",
				"error",
				err.Error(),
				"objectKey",
				objKey,
				"bucket",
				repo.bucket.bucketName,
				"metadataKey",
				repo.bucket.metadataKey,
			)
			return
		}
		if url == nil {
			repo.logger.Warn(
				ctx,
				"The url, that should be attached to the objects metadata was not found",
				"objectKey",
				objKey,
				"bucket",
				repo.bucket.bucketName,
				"metadataKey",
				repo.bucket.metadataKey,
			)
			u := ""
			url = &u
		}

		mapChan <- &domain.Map{
			Bounds:  *bounds,
			MapType: *mapType,
			Date:    *date,
			Source: domain.MapSource{
				MapProvider: *provider,
				URL:         *url,
			},
		}
	}
}

func createKeyFromMap(m domain.Map) string {
	d := m.Date
	s := fmt.Sprintf(
		"%s/%s/%s/%d_%d_%d.tif",
		m.Source.MapProvider.String(),
		m.Bounds.String(),
		m.MapType.String(),
		d.Year,
		d.Month,
		d.Day,
	)
	return s
}

func convertStringPathToDomain(objectKey string) (*domain.MapProvider, *domain.Bounds, *domain.MapType, error) {
	parts := strings.Split(objectKey, "/")
	if len(parts) != 3 {
		return nil, nil, nil, errors.New("the objectKey was not in the expected format")
	}
	objProvider, objBounds, objMapType := parts[0], parts[1], parts[2]
	provider := domain.StringToMapProvider(objProvider)
	bounds := domain.StringToBounds(objBounds)
	mapType := domain.StringToMapType(objMapType)
	return &provider, &bounds, &mapType, nil
}

func isDesiredObject(
	provider domain.MapProvider,
	desiredProvider domain.MapProvider,
	bounds domain.Bounds,
	desiredBounds domain.Bounds,
	mapType domain.MapType,
	desiredMapType domain.MapType,
) bool {
	return (desiredProvider == domain.MapProviderUnspecified || provider == desiredProvider) &&
		(desiredBounds == domain.BoundsUnspecified || bounds == desiredBounds) &&
		(desiredMapType == domain.MapTypeUnspecified || mapType == desiredMapType)
}

func sortMapsByDate(maps []domain.Map) []domain.Map {
	sort.Slice(maps, func(i, j int) bool {
		map1 := maps[i]
		map2 := maps[j]

		if map1.Date.Year != map2.Date.Year {
			return map1.Date.Year < map2.Date.Year
		}
		if map1.Date.Month != map2.Date.Month {
			return map1.Date.Month < map2.Date.Month
		}
		return map1.Date.Day < map2.Date.Day
	})
	return maps
}
