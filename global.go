package ylog

import (
	"context"
	"go.uber.org/zap/zapcore"
	"os"

	"go.uber.org/zap"
)

var globalLogger Logger

func getGlobalLogger() Logger {
	if globalLogger == nil {
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				TimeKey:        "ts",
				MessageKey:     "msg",
				EncodeDuration: zapcore.MillisDurationEncoder,
				EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
				LineEnding:     zapcore.DefaultLineEnding,
				LevelKey:       "level",
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
			}),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), // pipe to multiple writer
			zapcore.DebugLevel,
		)

		log := zap.New(core)
		return NewZap(log)
	}
	return globalLogger
}

func SetGlobalLogger(log Logger) {
	globalLogger = log
}

func Debug(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Error(ctx, msg, fields...)
}

func Access(ctx context.Context, data AccessLogData) {
	getGlobalLogger().Access(ctx, data)
}
