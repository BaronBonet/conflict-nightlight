package infrastructure

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

var CorrelationIDKey = GetEnvOrDefault("CORRELATION_ID_KEY", "correlation-id")

func NewContext() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, CorrelationIDKey, uuid.NewString())
}
func ExtractCorrelationIDFromContext(ctx context.Context) string {
	correlationId, ok := ctx.Value(CorrelationIDKey).(string)
	if !ok {
		return "unknown"
	}
	return correlationId
}

func AddCorrelationIdToContext(ctx context.Context, correlationId *string) (context.Context, error) {
	if correlationId == nil {
		return ctx, errors.New("correlationId is nil")
	}
	return context.WithValue(ctx, CorrelationIDKey, *correlationId), nil
}
