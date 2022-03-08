# ylog

ylog is an alternative logger for golang that support propagated data into log.


**Why data propagation?**

In most case when building a backend system, we frequently need data that exist in all log to make debugging in production easier.
For example, we want to add `trace_id` in every log that have same life-cycle (one API request-response), or
we want to add `journey_id` an ID from mobile application to ensure that 2 API calls is come from one journey 
(i.e: when we want to list the restaurant, we call 2 API one is get the list of restaurant that open and second is restaurant's details (this common in microservice)).

Some other log libraries like [logrus](https://github.com/sirupsen/logrus) or [zap](https://github.com/uber-go/zap) support `With` or `WithFields` 
method that add field into all log written after the function is called. But, context for propagation data is more convenient in Golang.

## How to use?

```go
func main() {
    type RequestPropagationData struct {
        TraceID string `tracer:"trace_id"`
    }

    propagateData := RequestPropagationData{
        TraceID: "abc",
    }
    
    tracer, err := ylog.NewTracer(propagateData, ylog.WithTag("tracer"))
    if err != nil {
		panic(err)
    }
    
    ctx := ylog.Inject(context.Background(), tracer)

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
	
    zapLogger := zap.New(core)
    uniLogger := ylog.NewZap(zapLogger)
	
	// set to global logger
	ylog.SetGlobalLogger(uniLogger)
    ylog.Info(ctx, "info message")
    ylog.Error(ctx, "error message")
}

```

Above code will return log output similar like this

```json
{"level":"info","ts":"2022-03-08T15:01:11.231938+07:00","msg":"info message","tag":"sys","tracer":{"trace_id":"abc"}}
{"level":"error","ts":"2022-03-08T15:01:11.23194+07:00","msg":"error message","tag":"sys","tracer":{"trace_id":"abc"}}

```

As you can see, the output always have `tracer.trace_id` data with exactly the same value. 
This value must always be generated in middleware as unique value, i.e: UUID.

If you want to get the data back, you can use:

```go
var wantDataBack RequestPropagationData
tracerData := ylog.MustExtract(ctx)
err := ylog.UnmarshalTracer(tracerData, &wantDataBack)
if err != nil {
	panic(err)
}

fmt.Println(wantDataBack)
```

Variable `wantDataBack` now will contain the same values as when you inject it to context. 

## Limitation

* Struct passed in `ylog.NewTracer` must only have field with simple type (string, number or boolean). 
  This also means that Field with type map, struct, or interface won't be supported.
  This is because the tracer data must contain simple value only, i.e: tracer_id, journey_id, user_id.
  Complex nested data requires to be nested (or even recursive) loop that may reduce performance of your system.

## Benchmark

```
go test -bench=. -benchmem ./...
goos: darwin
goarch: arm64
pkg: github.com/yusufsyaifudin/ylog
BenchmarkNewTracer-8     3994966               293.6 ns/op           536 B/op          7 allocs/op
BenchmarkNewZap-8         843906              1392 ns/op             857 B/op         11 allocs/op
PASS
ok      github.com/yusufsyaifudin/ylog  2.772s

```
