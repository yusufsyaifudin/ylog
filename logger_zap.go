package ylog

import (
	"context"

	"go.uber.org/zap"
)

type Zap struct {
	writer *zap.Logger
}

func NewZap(zapLogger *zap.Logger) *Zap {
	return &Zap{writer: zapLogger}
}

func (z *Zap) Debug(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Debug(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Info(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Info(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Warn(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Warn(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Error(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.With()
	z.writer.Error(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Access(ctx context.Context, data AccessLogData) {
	z.writer.Info(TypeAccessLog, localFieldZapFields(ctx, TypeAccessLog, []KeyValue{KV("data", data)})...)
}

func localFieldZapFields(ctx context.Context, tag string, fields []KeyValue) []zap.Field {
	zapFields := make([]zap.Field, 0)
	zapFields = append(zapFields, zap.String("tag", tag))

	data, ok := Extract(ctx)
	if ok {
		tracerData := make(map[string]interface{})
		for _, v := range data.values {
			tracerData[v.LogTagName] = v.Value
		}

		zapFields = append(zapFields, zap.Any("tracer", tracerData))
	}

	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.key, field.value))
	}

	return zapFields
}

var _ Logger = (*Zap)(nil)
