package internalmapsrepo

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestAWSMapsRepo_extractDateFromFilepath(t *testing.T) {
	d, _ := extractDateFromKey("2022_1_1.tif")
	assert.Equal(t, &domain.Date{
		Day:   1,
		Month: 1,
		Year:  2022,
	}, d)
}

func TestAWSMapsRepo_createFilepathFromDate(t *testing.T) {
	n := createKeyFromMap(domain.Map{MapType: domain.MapTypeDaily, Bounds: domain.BoundsUkraineAndAround, Source: domain.MapSource{
		MapProvider: domain.MapProviderEogdata,
		URL:         "test.com",
	}, Date: domain.Date{
		Day:   0,
		Month: 1,
		Year:  2022,
	}})
	assert.Equal(t, n, "MapProviderEogdata/BoundsUkraineAndAround/MapTypeDaily/2022_1_0.tif")
}

func TestAWSMapsRepo_List(t *testing.T) {
	ctx := context.Background()
	testBucketName := "test-bucket"
	testMetadataKey := "test-metadata-key"
	testObjects := []string{
		"MapProviderEogdata/BoundsUkraineAndAround/MapTypeDaily/2022_2_1.tif",
		"MapProviderEogdata/BoundsUkraineAndAround/MapTypeDaily/2022_1_1.tif",
		"MapProviderEogdata/BoundsUkraineAndAround/MapTypeDaily/2022_invalid_month_1.tif", // Malformed date
	}
	testMetadata := "https://example.com/test-url"

	mockLogger := ports.NewMockLogger(t)
	mockAWSClient := awsclient.NewMockAWSClient(t)

	testCases := []struct {
		name          string
		provider      domain.MapProvider
		bounds        domain.Bounds
		mapType       domain.MapType
		expectedMaps  []domain.Map
		expectedError error
	}{
		{
			name:     "List valid maps",
			provider: domain.MapProviderEogdata,
			bounds:   domain.BoundsUkraineAndAround,
			mapType:  domain.MapTypeDaily,
			expectedMaps: []domain.Map{
				{
					Bounds:  domain.BoundsUkraineAndAround,
					MapType: domain.MapTypeDaily,
					Date:    domain.Date{Day: 1, Month: 1, Year: 2022},
					Source: domain.MapSource{
						MapProvider: domain.MapProviderEogdata,
						URL:         testMetadata,
					},
				},
				{
					Bounds:  domain.BoundsUkraineAndAround,
					MapType: domain.MapTypeDaily,
					Date:    domain.Date{Day: 1, Month: 2, Year: 2022},
					Source: domain.MapSource{
						MapProvider: domain.MapProviderEogdata,
						URL:         testMetadata,
					},
				},
			},
		},
		{
			name:     "List valid maps when map type and bounds are not specified",
			provider: domain.MapProviderEogdata,
			bounds:   domain.BoundsUnspecified,
			mapType:  domain.MapTypeUnspecified,
			expectedMaps: []domain.Map{
				{
					Bounds:  domain.BoundsUkraineAndAround,
					MapType: domain.MapTypeDaily,
					Date:    domain.Date{Day: 1, Month: 1, Year: 2022},
					Source: domain.MapSource{
						MapProvider: domain.MapProviderEogdata,
						URL:         testMetadata,
					},
				},
				{
					Bounds:  domain.BoundsUkraineAndAround,
					MapType: domain.MapTypeDaily,
					Date:    domain.Date{Day: 1, Month: 2, Year: 2022},
					Source: domain.MapSource{
						MapProvider: domain.MapProviderEogdata,
						URL:         testMetadata,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger.On("Error", ctx, "There was an error when attempting to extract the date from the objectKey", "error", mock.AnythingOfType("string"), "objectKey", testObjects[2])

			mockAWSClient.On("ListObjectsInS3", ctx, testBucketName).Return(testObjects, nil)
			mockAWSClient.On("GetObjectMetadataInS3", ctx, testBucketName, mock.AnythingOfType("string"), testMetadataKey).Return(&testMetadata, nil)

			repo := NewAWSInternalMapsRepository(ctx, mockLogger, testBucketName, testMetadataKey, "test-queue", "/tmp")
			repo.awsClient = mockAWSClient

			maps, err := repo.List(ctx, tc.provider, tc.bounds, tc.mapType)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedMaps, maps)

			mockLogger.AssertCalled(t, "Error", ctx, "There was an error when attempting to extract the date from the objectKey", "error", mock.AnythingOfType("string"), "objectKey", testObjects[2])

			mockLogger.AssertExpectations(t)
			mockAWSClient.AssertExpectations(t)
		})
	}
}

func TestAWSMapsRepo_Download(t *testing.T) {
	ctx := context.Background()
	testBucketName := "test-bucket"
	testKey := "MapProviderEogdata/BoundsUkraineAndAround/MapTypeDaily/2022_1_1.tif"

	mockLogger := ports.NewMockLogger(t)
	mockAWSClient := awsclient.NewMockAWSClient(t)

	testMap := domain.Map{
		Bounds:  domain.BoundsUkraineAndAround,
		MapType: domain.MapTypeDaily,
		Date:    domain.Date{Day: 1, Month: 1, Year: 2022},
		Source: domain.MapSource{
			MapProvider: domain.MapProviderEogdata,
			URL:         "https://example.com/test-url",
		},
	}

	testData := []byte("test-data")

	mockAWSClient.On("GetFromS3", ctx, testBucketName, testKey).Return(testData, nil)

	repo := NewAWSInternalMapsRepository(ctx, mockLogger, testBucketName, "test-metadata-key", "test-queue", "/tmp")
	repo.awsClient = mockAWSClient

	localMap, err := repo.Download(ctx, testMap)
	assert.Nil(t, err)
	assert.NotNil(t, localMap)
	assert.Equal(t, testMap, localMap.Map)
	assert.FileExists(t, localMap.Filepath)

	mockAWSClient.AssertExpectations(t)
	err = os.Remove(localMap.Filepath)
}

func TestAWSMapsRepo_Create(t *testing.T) {
	ctx := context.Background()
	testQueueName := "test-queue"

	mockLogger := ports.NewMockLogger(t)
	mockAWSClient := awsclient.NewMockAWSClient(t)

	testMap := domain.Map{
		Bounds:  domain.BoundsUkraineAndAround,
		MapType: domain.MapTypeDaily,
		Date:    domain.Date{Day: 1, Month: 1, Year: 2022},
		Source: domain.MapSource{
			MapProvider: domain.MapProviderEogdata,
			URL:         "https://example.com/test-url",
		},
	}

	protoMap := prototransformers.DomainToProto(testMap)
	message := conflict_nightlightv1.RequestWrapper_DownloadAndCropRawTifRequest{DownloadAndCropRawTifRequest: &conflict_nightlightv1.DownloadAndCropRawTifRequest{Map: &protoMap}}

	mockAWSClient.On("PublishMessageToSQS", ctx, testQueueName, &message).Return(nil)

	repo := NewAWSInternalMapsRepository(ctx, mockLogger, "test-bucket", "test-metadata-key", testQueueName, "/tmp")
	repo.awsClient = mockAWSClient

	err := repo.Create(ctx, testMap)
	assert.NoError(t, err)

	mockAWSClient.AssertExpectations(t)
}
