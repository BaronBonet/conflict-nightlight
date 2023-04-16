package handlers

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/internalmapsrepository"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/productMapRepository"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/aws/aws-lambda-go/events"
	"google.golang.org/protobuf/encoding/protojson"
)

type MapPublisherLambdaEventHandler struct {
	logger ports.Logger
}

func NewMapPublisherLambdaHandler(logger ports.Logger) *MapPublisherLambdaEventHandler {
	return &MapPublisherLambdaEventHandler{logger: logger}
}

func (handler *MapPublisherLambdaEventHandler) HandleEvent(ctx context.Context, event events.SQSEvent) {

	for _, message := range event.Records {
		ctx = infrastructure.AddCorrelationIdToContext(ctx, message.MessageAttributes[infrastructure.CorrelationIDKey].StringValue)

		request := conflict_nightlightv1.RequestWrapper{}
		err := protojson.Unmarshal([]byte(message.Body), &request)
		if err != nil {
			handler.logger.Fatal(ctx, "Error while unmarshalling json for publish map request.", "error", err,
				"event", message)
		}

		mapBoxRepo := productMapRepository.NewMapboxRepo(ctx, handler.logger, infrastructure.GetEnvOrDefault("CONFLICT_NIGHTLIGHT_SECRETS_KEY", "conflict-nightlight-secrets"), productMapRepository.JsonForFrontend{BucketName: infrastructure.GetEnvOrDefault("CDN_BUCKET_NAME", "conflict-nightlight-cdn"), ObjectKey: infrastructure.GetEnvOrDefault("FRONTEND_MAP_OPTIONS_JSON", "conflict-nightlight-map-options.json")})
		internalRepo := internalmapsrepository.NewAWSInternalMapsRepository(ctx, handler.logger, infrastructure.GetEnvOrDefault("PROCESSED_TIF_BUCKET_NAME", "conflict-nightlight-processed-tif"), infrastructure.GetEnvOrDefault("SOURCE_KEY_URL", "source-url"), "", infrastructure.GetEnvOrDefault("WRITE_DIR", "/tmp"))
		service := services.NewProductService(handler.logger, internalRepo, mapBoxRepo)

		service.PublishMap(ctx, prototransformers.ProtoToDomain(request.GetPublishMapProductRequest().Map))
	}
}
