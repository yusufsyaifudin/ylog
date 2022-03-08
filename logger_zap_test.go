package ylog_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yusufsyaifudin/ylog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BenchmarkNewTracer(b *testing.B) {
	type Tracer struct {
		RemoteIP string
	}

	tracer := Tracer{
		RemoteIP: "123",
	}

	for i := 0; i < b.N; i++ {
		tr, err := ylog.NewTracer(tracer)
		if err != nil {
			b.Fatal(err)
			return
		}

		var actual Tracer
		err = ylog.UnmarshalTracer(tr, &actual)
		if err != nil {
			b.Fatal(err)
			return
		}

		//fmt.Println(tracer)
		//fmt.Println(actual)
	}

}

func BenchmarkNewZap(b *testing.B) {
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
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(io.Discard)), // pipe to multiple writer
		zapcore.DebugLevel,
	)
	zapLogger := zap.New(core)
	uniLogger := ylog.NewZap(zapLogger)

	type RequestPropagationData struct {
		TraceID string `tracer:"trace_id"`
	}

	propagateData := RequestPropagationData{
		TraceID: time.Now().String(),
	}

	tracer, err := ylog.NewTracer(propagateData, ylog.WithTag("tracer"))
	assert.NotNil(b, tracer)
	assert.NoError(b, err)

	ctx := ylog.Inject(context.Background(), tracer)
	for i := 0; i < b.N; i++ {
		// zapLogger.Error("message", zap.Any("tracer", logger.Tracer{AppTraceID: "test"}))
		uniLogger.Error(ctx, "message")
	}

}
