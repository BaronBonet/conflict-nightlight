package ports

import (
	"context"
)

// Logger is a special port that belongs to the service and the adapters
type Logger interface {
	Debug(ctx context.Context, msg string, keysAndValues ...interface{})
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
	Fatal(ctx context.Context, msg string, keysAndValues ...interface{})
}
