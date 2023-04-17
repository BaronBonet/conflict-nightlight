package main

import (
	"context"
	"fmt"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/externalmapsrepo"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/frontendmapdatarepo"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/internalmapsrepo"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/maptileserverrepo"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/handlers"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	request := conflict_nightlightv1.RequestWrapper{Message: &conflict_nightlightv1.RequestWrapper_PublishMapProductRequest{PublishMapProductRequest: &conflict_nightlightv1.PublishMapProductRequest{Map: &conflict_nightlightv1.Map{
		Date: &conflict_nightlightv1.Date{
			Day:   0,
			Month: 1,
			Year:  2021,
		},
		MapType: conflict_nightlightv1.MapType(domain.StringToMapType("Monthly")),
		Bounds:  1,
		MapSource: &conflict_nightlightv1.MapSource{
			MapProvider: 1,
			Url:         "example.com",
		},
	}}}}
	v, err := protojson.Marshal(&request)
	if err != nil {
		panic("cannot marshal")
	}

	s := "1234"
	event := events.SQSEvent{Records: []events.SQSMessage{{
		Body: string(v),
		MessageAttributes: map[string]events.SQSMessageAttribute{"correlation-id": {
			StringValue: &s,
			DataType:    "string",
		},
		}}}}
	ctx := context.Background()
	logger := adapters.NewZapLogger(zap.NewDevelopmentConfig(), true)
	writeDir := infrastructure.GetEnvOrDefault("WRITE_DIR", "/tmp")
	externalMapsRepo := externalmapsrepo.NewEogdataExternalMapsRepository(logger, fmt.Sprintf("%s/eog_cache", writeDir))
	internalRawMapsRepo := internalmapsrepo.NewAWSInternalMapsRepository(ctx,
		logger,
		infrastructure.GetEnvOrDefault("RAW_TIF_BUCKET", "conflict-nightlight-raw-tif"),
		infrastructure.GetEnvOrDefault("SOURCE_URL_KEY", "source-url"),
		infrastructure.GetEnvOrDefault("DOWNLOAD_RAW_TIF_QUEUE", "conflict-nightlight-download-and-crop-raw-tif-request"),
		writeDir,
	)
	internalProcessedMapsRepo := internalmapsrepo.NewAWSInternalMapsRepository(ctx, logger, infrastructure.GetEnvOrDefault("PROCESSED_TIF_BUCKET_NAME", "conflict-nightlight-processed-tif"), infrastructure.GetEnvOrDefault("SOURCE_KEY_URL", "source-url"), "", infrastructure.GetEnvOrDefault("WRITE_DIR", "/tmp"))
	frontendMapDataRepo := frontendmapdatarepo.NewS3FrontendMapDataRepo(ctx, logger, infrastructure.GetEnvOrDefault("CDN_BUCKET_NAME", "conflict-nightlight-cdn"), infrastructure.GetEnvOrDefault("FRONTEND_MAP_OPTIONS_JSON", "conflict-nightlight-map-options.json"))
	mapboxTileServerRepo := maptileserverrepo.NewMapboxTileServerRepo(ctx, logger, infrastructure.GetEnvOrDefault("CONFLICT_NIGHTLIGHT_SECRETS_KEY", "conflict-nightlight-secrets"))
	service := services.NewOrchestratorService(logger, externalMapsRepo, internalRawMapsRepo, internalProcessedMapsRepo, frontendMapDataRepo, mapboxTileServerRepo)

	handler := handlers.NewMapPublisherLambdaHandler(logger, service)
	handler.HandleEvent(context.TODO(), event)
}
