package internalmapsrepository

import (
	"context"
	"errors"
	"fmt"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AWSMapsRepo struct {
	logger       ports.Logger
	s3Repo       s3Bucket
	messageQueue persistRequestQueue
	tmpWriteDir  string
}

type persistRequestQueue struct {
	client sqs.Client
	name   string
}

type s3Bucket struct {
	client      *s3.Client
	bucketName  string
	metadataKey string
}

func NewAWSInternalMapsRepository(ctx context.Context, logger ports.Logger, repoBucketName string, metadataKey string, persistRequestQueueName string, tmpWriteDir string) *AWSMapsRepo {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatal(ctx, "Error when attempting to load the aws config", "error", err)
	}
	return &AWSMapsRepo{logger: logger, s3Repo: s3Bucket{
		client:      s3.NewFromConfig(cfg),
		bucketName:  repoBucketName,
		metadataKey: metadataKey,
	}, messageQueue: persistRequestQueue{
		client: *sqs.NewFromConfig(cfg),
		name:   persistRequestQueueName,
	}, tmpWriteDir: tmpWriteDir}
}

func (repo *AWSMapsRepo) List(ctx context.Context, provider domain.MapProvider, bounds domain.Bounds, mapType domain.MapType) ([]domain.Map, error) {
	result, err := repo.s3Repo.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(repo.s3Repo.bucketName),
	})

	if err != nil {
		repo.logger.Error(ctx, "Couldn't list objects in bucket.", "bucket-name", repo.s3Repo.bucketName, "error", err)
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	bucketDirectory := fmt.Sprintf("%s/%s/%s", provider.String(), bounds.String(), mapType.String())

	var existingMaps []domain.Map
	for _, item := range result.Contents {
		objectKey := item.Key
		directory := filepath.Dir(*objectKey)
		extension := filepath.Ext(*objectKey)

		if directory == bucketDirectory && extension == ".tif" {
			date, err := extractDateFromFilepath(*objectKey)
			if err != nil {
				repo.logger.Error(ctx, "There was an error when attempting to extract the date from the objectKey", "error", err.Error(), "objectKey", objectKey)
			}
			existingMaps = append(existingMaps, domain.Map{
				Bounds:  bounds,
				MapType: mapType,
				Date:    *date,
				Source: domain.MapSource{
					MapProvider: provider,
					URL:         *repo.getMetadata(ctx, objectKey),
				},
			})
		}
	}
	return existingMaps, nil
}

func (repo *AWSMapsRepo) Download(ctx context.Context, m domain.Map) *domain.LocalMap {
	key := createFilepathFromMap(m)
	localFilepath := fmt.Sprintf("%s/%s.tif", repo.tmpWriteDir, uuid.NewString())
	result, err := repo.s3Repo.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(repo.s3Repo.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		repo.logger.Fatal(ctx, "Couldn't download file", "key", "bucket", repo.s3Repo.bucketName, key, "error", err)
	}
	defer result.Body.Close()
	file, err := os.Create(localFilepath)
	if err != nil {
		repo.logger.Fatal(ctx, "Couldn't create file", "localFilepath", localFilepath, "error", err)
	}
	defer file.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		repo.logger.Fatal(ctx, "Couldn't read object body.", "key", key, "error", err)
	}
	_, err = file.Write(body)
	if err != nil {
		repo.logger.Fatal(ctx, "Couldn't write contents from s3", "key", key, "error", err)
	}
	return &domain.LocalMap{
		Filepath: localFilepath,
		Map:      m,
	}
}

// Create puts a message on a sqs queue which will be picked up by a python application responsible for downloading the tif file
func (repo *AWSMapsRepo) Create(ctx context.Context, m domain.Map) error {
	correlationID := infrastructure.ExtractCorrelationIDFromContext(ctx)
	dataType := "String"
	protoMap := prototransformers.DomainToProto(m)
	message := conflict_nightlightv1.RequestWrapper{Message: &conflict_nightlightv1.RequestWrapper_DownloadAndCropRawTifRequest{DownloadAndCropRawTifRequest: &conflict_nightlightv1.DownloadAndCropRawTifRequest{Map: &protoMap}}}
	smInput := &sqs.SendMessageInput{
		DelaySeconds: 0,
		MessageBody:  aws.String(protojson.Format(&message)),
		QueueUrl:     aws.String(repo.messageQueue.name),
		MessageAttributes: map[string]types.MessageAttributeValue{infrastructure.CorrelationIDKey: {
			DataType:    &dataType,
			StringValue: &correlationID,
		}},
	}
	_, err := repo.messageQueue.client.SendMessage(ctx, smInput)
	return err
}

func (repo *AWSMapsRepo) getMetadata(ctx context.Context, objectKey *string) *string {
	objectInfo, err := repo.s3Repo.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(repo.s3Repo.bucketName),
		Key:    aws.String(*objectKey),
	})
	if err != nil {
		repo.logger.Error(ctx, "There was an error when attempting to extract the objects metadata", "error", err.Error())
		return nil
	}
	link := objectInfo.Metadata[repo.s3Repo.metadataKey]
	if link == "" {
		repo.logger.Error(ctx, "There was no metadata key on this object", "objectKey", objectKey, "expected-key", repo.s3Repo.metadataKey)
		return nil
	}
	return &link
}

func extractDateFromFilepath(filePath string) (*domain.Date, error) {
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

	return &domain.Date{
		Day:   0,
		Month: month,
		Year:  year,
	}, nil
}

func createFilepathFromMap(m domain.Map) string {
	d := m.Date
	s := fmt.Sprintf("%s/%s/%s/%d_%d_%d.tif", m.Source.MapProvider.String(), m.Bounds.String(), m.MapType.String(), d.Year, d.Month, d.Day)
	return s
}
