package handlers

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/aws/aws-lambda-go/events"
	"google.golang.org/protobuf/encoding/protojson"
)

type MapPublisherLambdaEventHandler struct {
	logger ports.Logger
	srv    ports.OrchestratorService
}

func NewMapPublisherLambdaHandler(logger ports.Logger, srv ports.OrchestratorService) *MapPublisherLambdaEventHandler {
	return &MapPublisherLambdaEventHandler{logger: logger, srv: srv}
}

func (handler *MapPublisherLambdaEventHandler) HandleEvent(ctx context.Context, event events.SQSEvent) {
	for _, message := range event.Records {
		handler.logger.Debug(ctx, "Received message from queue.", "message", message, "messageAttributes", message.MessageAttributes)
		ctx, err := infrastructure.AddCorrelationIdToContext(ctx, message.MessageAttributes[infrastructure.CorrelationIDKey].StringValue)
		if err != nil {
			handler.logger.Warn(ctx, "Error while adding correlation id to context.", "error", err)
		}
		request := conflict_nightlightv1.RequestWrapper{}
		handler.logger.Info(ctx, "Unmarshalling json for publish map request.", "message", message)
		if err := protojson.Unmarshal([]byte(message.Body), &request); err != nil {
			handler.logger.Fatal(ctx, "Error while unmarshalling json for publish map request.", "error", err,
				"event", message)
		}
		m := prototransformers.ProtoToDomain(request.GetPublishMapProductRequest().Map)
		if err := handler.srv.PublishMap(ctx, m); err != nil {
			handler.logger.Fatal(ctx, "Error while publishing map.", "error", err)
		}
		handler.logger.Info(ctx, "Map published successfully.", "map", m)
	}
}
