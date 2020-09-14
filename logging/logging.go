package logging

import (
	"context"

	"bitbucket.org/kleinnic74/photos/consts"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerKeyType string

const (
	loggerKey = loggerKeyType("logger")
)

var rootLogger *zap.Logger

func init() {
	var err error
	devmode := consts.IsDevMode()
	if devmode {
		rootLogger, err = zap.NewDevelopment()
	} else {
		rootLogger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	rootLogger.With(zap.Bool("devmode", devmode)).Info("Logging initialized")
}

func From(ctx context.Context) *zap.Logger {
	l := ctx.Value(loggerKey)
	if l == nil {
		return rootLogger
	}
	return l.(*zap.Logger)
}

func SubFrom(ctx context.Context, name string) (*zap.Logger, context.Context) {
	logger := From(ctx).Named(name)
	return logger, Context(ctx, logger)
}

func Context(ctx context.Context, logger *zap.Logger) context.Context {
	if logger == nil {
		logger = rootLogger
	}
	return context.WithValue(ctx, loggerKey, logger)
}

func FromWithNameAndFields(ctx context.Context, name string, fields ...zapcore.Field) (*zap.Logger, context.Context) {
	logger := From(ctx).With(fields...).Named(name)
	ctx = Context(ctx, logger)
	return logger, ctx
}

func FromWithFields(ctx context.Context, fields ...zapcore.Field) (*zap.Logger, context.Context) {
	logger := From(ctx).With(fields...)
	ctx = Context(ctx, logger)
	return logger, ctx
}
