package logging

import (
	"context"
	"fmt"
	"os"

	"bitbucket.org/kleinnic74/photos/consts"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerKeyType string

const (
	loggerKey = loggerKeyType("logger")
	errorLog  = false
)

var rootLogger *zap.Logger

func init() {
	devmode := consts.IsDevMode()
	debugFilter := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})
	infoFilter := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})
	warnOrErrorFilter := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})

	var jsonEncoder zapcore.Encoder
	if devmode {
		jsonEncoder = zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		jsonEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}
	var cores []zapcore.Core
	if errorLog {
		errorFile, err := os.OpenFile("errors.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic(fmt.Sprintf("Failed to open log file: %s", err))
		}
		cores = append(cores, zapcore.NewCore(jsonEncoder, errorFile, warnOrErrorFilter))
	}
	var fileFilter zap.LevelEnablerFunc
	if devmode {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		console := zapcore.Lock(os.Stdout)
		cores = append(cores, zapcore.NewCore(consoleEncoder, console, debugFilter))
		fileFilter = debugFilter
	} else {
		fileFilter = infoFilter
	}

	logfile, err := os.OpenFile("log.json", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file: %s", err))
	}
	cores = append(cores, zapcore.NewCore(jsonEncoder, zapcore.Lock(logfile), fileFilter))

	// Loggly
	loggly := NewLogglySink("e8e5b948-f0d9-4c99-ae02-09f17a48d7ac")
	cores = append(cores, zapcore.NewCore(NewLogglyEncoder(), loggly, fileFilter))

	rootLogger = zap.New(zapcore.NewTee(cores...))
	rootLogger.With(zap.Bool("devmode", devmode)).Info("Logging initialized")
}

// From returns the logger of the current context, if no logger is available, returns the root logger
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
