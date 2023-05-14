package main

import (
	"context"
	"fmt"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/awsclient"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/externalmapsrepo"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/frontendmapdatarepo"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/internalmapsrepo"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/maptileserverrepo"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/handlers"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"
)

func main() {
	logger := adapters.NewZapLogger(zap.NewProductionConfig(), false)
	ctx := context.Background()
	writeDir := infrastructure.GetEnvOrDefault("WRITE_DIR", "/tmp")
	externalMapsRepo := externalmapsrepo.NewEogdataExternalMapsRepository(logger, fmt.Sprintf("%s/eog_cache", writeDir))
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
	if err != nil {
		logger.Fatal(ctx, "Error when attempting to load the aws config", "error", err)
	}
	awsClient := awsclient.NewAWSClient(cfg)
	internalRawMapsRepo := internalmapsrepo.NewAWSInternalMapsRepository(
		logger,
		infrastructure.GetEnvOrDefault("RAW_TIF_BUCKET", "conflict-nightlight-raw-tif"),
		infrastructure.GetEnvOrDefault("SOURCE_URL_KEY", "source-url"),
		infrastructure.GetEnvOrDefault("DOWNLOAD_RAW_TIF_QUEUE", "conflict-nightlight-download-and-crop-raw-tif-request"),
		writeDir,
		awsClient,
	)
	internalProcessedMapsRepo := internalmapsrepo.NewAWSInternalMapsRepository(logger, infrastructure.GetEnvOrDefault("PROCESSED_TIF_BUCKET_NAME", "conflict-nightlight-processed-tif"), infrastructure.GetEnvOrDefault("SOURCE_KEY_URL", "source-url"), "", infrastructure.GetEnvOrDefault("WRITE_DIR", "/tmp"), awsClient)
	frontendMapDataRepo := frontendmapdatarepo.NewS3FrontendMapDataRepo(ctx, logger, infrastructure.GetEnvOrDefault("CDN_BUCKET_NAME", "conflict-nightlight-cdn"), infrastructure.GetEnvOrDefault("FRONTEND_MAP_OPTIONS_JSON", "conflict-nightlight-map-options.json"))
	mapboxTileServerRepo := maptileserverrepo.NewMapboxTileServerRepo(ctx, logger, infrastructure.GetEnvOrDefault("CONFLICT_NIGHTLIGHT_SECRETS_KEY", "conflict-nightlight-secrets"))
	service := services.NewOrchestratorService(logger, externalMapsRepo, internalRawMapsRepo, internalProcessedMapsRepo, frontendMapDataRepo, mapboxTileServerRepo)
	lambdaHandler := handlers.NewMapPublisherLambdaHandler(logger, service)
	lambda.Start(lambdaHandler.HandleEvent)
}
