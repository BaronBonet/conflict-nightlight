package adapters

import (
	"context"

	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(config zap.Config, useDebug bool) ports.Logger {

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	log, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic("Cannot create logger: " + err.Error())
	}
	log = log.WithOptions(zap.AddCallerSkip(2))

	if useDebug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	return &zapLogger{
		logger: log.With(zap.String("version", infrastructure.Version)),
	}
}

func (l *zapLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, zap.DebugLevel, msg, keysAndValues...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, zap.InfoLevel, msg, keysAndValues...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, zap.WarnLevel, msg, keysAndValues...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, zap.ErrorLevel, msg, keysAndValues...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, zap.ErrorLevel, msg, keysAndValues...)
	panic(msg)
}

func (l *zapLogger) log(ctx context.Context, level zapcore.Level, msg string, keysAndValues ...interface{}) {
	if l.logger.Check(level, msg) == nil {
		return
	}
	fields := createFields(keysAndValues)
	logger := l.addCorrelationID(ctx)
	logger.Log(level, msg, fields...)
}

func (l *zapLogger) addCorrelationID(ctx context.Context) *zap.Logger {
	correlationID := infrastructure.ExtractCorrelationIDFromContext(ctx)
	return l.logger.With(
		zap.String(infrastructure.CorrelationIDKey, correlationID))
}

func createFields(values []interface{}) []zap.Field {
	fields := make([]zap.Field, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		fields[i/2] = zap.Any(values[i].(string), values[i+1])
	}
	return fields
}
