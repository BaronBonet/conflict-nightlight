package frontendmapdatarepo

import (
	"bytes"
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestS3FrontendMapDataRepo_Update(t *testing.T) {
	ctx := context.Background()
	mockLogger := ports.NewMockLogger(t)
	mockAWSClient := awsclient.NewMockAWSClient(t)
	mapRepo := &s3FrontendMapDataRepo{logger: mockLogger, awsClient: mockAWSClient, bucketName: "test-bucket", objectKey: "test-key"}

	url := "mapbox://test-tileset"
	testMap := domain.PublishedMap{Map: domain.Map{
		Date:    domain.Date{Day: 1, Month: 2, Year: 2023},
		MapType: domain.MapTypeDaily,
		Bounds:  domain.BoundsUkraineAndAround,
		Source: domain.MapSource{
			MapProvider: domain.MapProviderEogdata,
			URL:         "mapbox://test-tileset",
		},
	}, Url: url}

	t.Run("JSON not found in S3 bucket", func(t *testing.T) {
		noSuchKeyError := &types.NoSuchKey{}

		mockAWSClient.On("GetFromS3", ctx, "test-bucket", "test-key").Return([]byte{}, noSuchKeyError)
		mockAWSClient.On("UploadToS3", ctx, "test-bucket", "test-key", mock.Anything).Return(nil)

		mapRepo.Upsert(ctx, testMap)

		mockAWSClient.AssertExpectations(t)
	})

	t.Run("JSON found in S3 bucket and updated", func(t *testing.T) {
		existingJSON := `[{"display_name":"Jan 2023","url":"mapbox://existing-tileset","key":"Daily-UkraineAnd_2023-1-1"}]`
		expectedJSON := `[{"display_name":"Jan 2023","url":"mapbox://existing-tileset","key":"Daily-UkraineAnd_2023-1-1"},{"display_name":"Feb 2023","url":"mapbox://test-tileset","key":"Daily-UkraineAnd_2023-2-1"}]`

		mockAWSClient.On("GetFromS3", ctx, "test-bucket", "test-key").Return([]byte(existingJSON), nil)
		mockAWSClient.On("UploadToS3", ctx, "test-bucket", "test-key", mock.AnythingOfType("*bytes.Reader")).Run(func(args mock.Arguments) {
			body := args.Get(3).(*bytes.Reader)
			buf := new(bytes.Buffer)
			buf.ReadFrom(body)
			assert.JSONEq(t, expectedJSON, buf.String())
		}).Return(nil)

		mapRepo.Upsert(ctx, testMap)

		mockAWSClient.AssertExpectations(t)
	})
}

func TestS3FrontendMapDataRepo_Delete(t *testing.T) {
	ctx := context.Background()
	mockLogger := ports.NewMockLogger(t)
	mockAWSClient := awsclient.NewMockAWSClient(t)
	mapRepo := &s3FrontendMapDataRepo{logger: mockLogger, awsClient: mockAWSClient, bucketName: "test-bucket", objectKey: "test-key"}

	testMap := domain.Map{
		Date:    domain.Date{Day: 1, Month: 2, Year: 2023},
		MapType: domain.MapTypeDaily,
		Bounds:  domain.BoundsUkraineAndAround,
		Source: domain.MapSource{
			MapProvider: domain.MapProviderEogdata,
			URL:         "mapbox://test-tileset",
		},
	}

	t.Run("JSON found in S3 bucket and map deleted", func(t *testing.T) {
		existingJSON := `[{"display_name":"Jan 2023","url":"mapbox://existing-tileset","key":"Daily-UkraineAnd_2023-1-1"},{"display_name":"Feb 2023","url":"mapbox://test-tileset","key":"Daily-UkraineAnd_2023-2-1"}]`
		expectedJSON := `[{"display_name":"Jan 2023","url":"mapbox://existing-tileset","key":"Daily-UkraineAnd_2023-1-1"}]`

		mockAWSClient.On("GetFromS3", ctx, "test-bucket", "test-key").Return([]byte(existingJSON), nil)
		mockAWSClient.On("UploadToS3", ctx, "test-bucket", "test-key", mock.AnythingOfType("*bytes.Reader")).Run(func(args mock.Arguments) {
			body := args.Get(3).(*bytes.Reader)
			buf := new(bytes.Buffer)
			buf.ReadFrom(body)
			assert.JSONEq(t, expectedJSON, buf.String())
		}).Return(nil)

		err := mapRepo.Delete(ctx, testMap)
		assert.NoError(t, err)
		mockAWSClient.AssertExpectations(t)
	})

	t.Run("JSON found in S3 bucket but map not found", func(t *testing.T) {
		existingJSON := `[{"display_name":"Jan 2023","url":"mapbox://existing-tileset","key":"Daily-UkraineAnd_2023-1-1"}]`

		mockAWSClient.On("GetFromS3", ctx, "test-bucket", "test-key").Return([]byte(existingJSON), nil)
		mockAWSClient.AssertNotCalled(t, "UploadToS3")

		err := mapRepo.Delete(ctx, testMap)
		assert.NoError(t, err)
		mockAWSClient.AssertExpectations(t)
	})
}

func TestS3FrontendMapDataRepo_List(t *testing.T) {
	ctx := context.Background()
	mockLogger := ports.NewMockLogger(t)
	mockAWSClient := awsclient.NewMockAWSClient(t)
	mapRepo := &s3FrontendMapDataRepo{logger: mockLogger, awsClient: mockAWSClient, bucketName: "test-bucket", objectKey: "test-key"}

	mapOptionsJSON := `[{
    "display_name": "Jan 2021",
    "url": "mapbox://ericcbonet.Monthly-UkraineAnd_2021-1-1",
    "key": "Monthly-UkraineAnd_2021-1-1",
    "map": {
      "date": {
        "day": 1,
        "month": 1,
        "year": 2021
      },
      "map_type": 2,
      "bounds": 1,
      "map_source": {
        "map_provider": 1,
        "url": "https://eogdata.mines.edu/nighttime_light/monthly/v10/2021/202101/vcmcfg/SVDNB_npp_20210101-20210131_75N060W_vcmcfg_v10_c202102062300.tgz"
      }
    }
  },
  {
    "display_name": "Feb 2021",
    "url": "mapbox://ericcbonet.Monthly-UkraineAnd_2021-2-1",
    "key": "Monthly-UkraineAnd_2021-2-1",
    "map": {
      "date": {
        "day": 1,
        "month": 2,
        "year": 2021
      },
      "map_type": 2,
      "bounds": 1,
      "map_source": {
        "map_provider": 1,
        "url": "https://eogdata.mines.edu/nighttime_light/monthly/v10/2021/202102/vcmcfg/SVDNB_npp_20210201-20210228_75N060W_vcmcfg_v10_c202103091200.tgz"
      }
    }
  }]`

	t.Run("List maps from S3 bucket", func(t *testing.T) {
		mockAWSClient.On("GetFromS3", ctx, "test-bucket", "test-key").Return([]byte(mapOptionsJSON), nil)

		maps, err := mapRepo.List(ctx)
		assert.NoError(t, err)

		expectedMaps := []domain.PublishedMap{{
			Map: domain.Map{Date: domain.Date{
				Day:   1,
				Month: 1,
				Year:  2021,
			}, MapType: domain.MapTypeMonthly, Bounds: domain.BoundsUkraineAndAround, Source: domain.MapSource{MapProvider: domain.MapProviderEogdata, URL: "https://eogdata.mines.edu/nighttime_light/monthly/v10/2021/202101/vcmcfg/SVDNB_npp_20210101-20210131_75N060W_vcmcfg_v10_c202102062300.tgz"}},
			Url: "mapbox://ericcbonet.Monthly-UkraineAnd_2021-1-1",
		}, {
			Map: domain.Map{Date: domain.Date{
				Day:   1,
				Month: 2,
				Year:  2021,
			}, MapType: domain.MapTypeMonthly, Bounds: domain.BoundsUkraineAndAround, Source: domain.MapSource{MapProvider: domain.MapProviderEogdata, URL: "https://eogdata.mines.edu/nighttime_light/monthly/v10/2021/202102/vcmcfg/SVDNB_npp_20210201-20210228_75N060W_vcmcfg_v10_c202103091200.tgz"}},
			Url: "mapbox://ericcbonet.Monthly-UkraineAnd_2021-2-1",
		}}
		assert.Equal(t, expectedMaps, maps)

		mockAWSClient.AssertExpectations(t)
	})
}

func TestS3FrontendMapDataRepo_updateMapOptionsList(t *testing.T) {
	testCases := []*struct {
		name           string
		mapOptionsList []*conflict_nightlightv1.MapOptions
		newOption      conflict_nightlightv1.MapOptions
		expected       []*conflict_nightlightv1.MapOptions
	}{
		{
			name: "Add new option",
			mapOptionsList: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "Jan 2021", Url: "mapbox://example1", Key: "Monthly-UkraineAnd_2021-1-1"},
				{DisplayName: "Feb 2021", Url: "mapbox://example2", Key: "Monthly-UkraineAnd_2021-2-1"},
			},
			newOption: conflict_nightlightv1.MapOptions{DisplayName: "Mar 2021", Url: "mapbox://example3", Key: "Monthly-UkraineAnd_2021-3-1"},
			expected: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "Jan 2021", Url: "mapbox://example1", Key: "Monthly-UkraineAnd_2021-1-1"},
				{DisplayName: "Feb 2021", Url: "mapbox://example2", Key: "Monthly-UkraineAnd_2021-2-1"},
				{DisplayName: "Mar 2021", Url: "mapbox://example3", Key: "Monthly-UkraineAnd_2021-3-1"},
			},
		},
		{
			name: "Upsert existing option",
			mapOptionsList: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "Jan 2021", Url: "mapbox://example1", Key: "Monthly-UkraineAnd_2021-1-1"},
				{DisplayName: "Feb 2021", Url: "mapbox://example2", Key: "Monthly-UkraineAnd_2021-2-1"},
			},
			newOption: conflict_nightlightv1.MapOptions{DisplayName: "Feb 2021", Url: "mapbox://example3", Key: "Monthly-UkraineAnd_2021-2-1"},
			expected: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "Jan 2021", Url: "mapbox://example1", Key: "Monthly-UkraineAnd_2021-1-1"},
				{DisplayName: "Feb 2021", Url: "mapbox://example3", Key: "Monthly-UkraineAnd_2021-2-1"},
			},
		},
		{
			name: "Add new option and sort with unsorted input",
			mapOptionsList: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "Jan 2021", Url: "mapbox://example1", Key: "Monthly-UkraineAnd_2021-1-1"},
				{DisplayName: "Mar 2021", Url: "mapbox://example3", Key: "Monthly-UkraineAnd_2021-3-1"},
			},
			newOption: conflict_nightlightv1.MapOptions{DisplayName: "Feb 2021", Url: "mapbox://example2", Key: "Monthly-UkraineAnd_2021-2-1"},
			expected: []*conflict_nightlightv1.MapOptions{
				{DisplayName: "Jan 2021", Url: "mapbox://example1", Key: "Monthly-UkraineAnd_2021-1-1"},
				{DisplayName: "Feb 2021", Url: "mapbox://example2", Key: "Monthly-UkraineAnd_2021-2-1"},
				{DisplayName: "Mar 2021", Url: "mapbox://example3", Key: "Monthly-UkraineAnd_2021-3-1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := updateMapOptionsList(tc.mapOptionsList, &tc.newOption)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
