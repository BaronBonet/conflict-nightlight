package awsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

//go:generate mockery --name=AWSClient
type AWSClient interface {
	UploadToS3(ctx context.Context, bucket, key string, data io.Reader) error
	GetFromS3(ctx context.Context, bucket, key string) ([]byte, error)
	DeleteFromS3(ctx context.Context, bucket, key string) error
	ListObjectsInS3(ctx context.Context, bucket string) ([]string, error)
	GetObjectMetadataInS3(ctx context.Context, bucket, key, metadataKey string) (*string, error)
	GetSecretFromSecretsManager(ctx context.Context, secretKey string) (secrets interface{}, err error)
	PublishMessageToSQS(ctx context.Context, queueName string, message interface{}) error
}

//go:generate mockery --name=SecretsManagerClientInterface
type SecretsManagerClientInterface interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type awsClient struct {
	s3Client             *s3.Client
	secretsManagerClient SecretsManagerClientInterface
	sqsClient            *sqs.Client
}

func NewAWSClient(cfg aws.Config) AWSClient {
	return &awsClient{
		s3Client:             s3.NewFromConfig(cfg),
		secretsManagerClient: secretsmanager.NewFromConfig(cfg),
		sqsClient:            sqs.NewFromConfig(cfg),
	}
}

func (c *awsClient) UploadToS3(ctx context.Context, bucket, key string, data io.Reader) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	return err
}

func (c *awsClient) GetFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *awsClient) DeleteFromS3(ctx context.Context, bucket, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err
}

func (c *awsClient) ListObjectsInS3(ctx context.Context, bucket string) ([]string, error) {
	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}

	var objectKeys []string
	for _, item := range result.Contents {
		objectKeys = append(objectKeys, *item.Key)
	}

	return objectKeys, nil
}

func (c *awsClient) GetObjectMetadataInS3(ctx context.Context, bucket, key, metadataKey string) (*string, error) {
	objectInfo, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	link := objectInfo.Metadata[metadataKey]
	if link == "" {
		return nil, fmt.Errorf("metadata key not found on the object: %s", metadataKey)
	}

	return &link, nil
}
func (c *awsClient) GetSecretFromSecretsManager(ctx context.Context, secretKey string) (interface{}, error) {
	value, err := c.secretsManagerClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretKey,
	})
	if err != nil {
		return nil, err
	}

	var unmarshalledData map[string]interface{}
	if err := json.Unmarshal([]byte(*value.SecretString), &unmarshalledData); err != nil {
		return nil, err
	}
	return unmarshalledData, nil
}

func (c *awsClient) PublishMessageToSQS(ctx context.Context, queueName string, message interface{}) error {
	correlationID := infrastructure.ExtractCorrelationIDFromContext(ctx)
	dataType := "String"
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	smInput := &sqs.SendMessageInput{
		DelaySeconds: 0,
		MessageBody:  aws.String(string(messageJSON)),
		QueueUrl:     aws.String(queueName),
		MessageAttributes: map[string]types.MessageAttributeValue{infrastructure.CorrelationIDKey: {
			DataType:    &dataType,
			StringValue: &correlationID,
		}},
	}

	_, err = c.sqsClient.SendMessage(ctx, smInput)
	return err
}
