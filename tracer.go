package ylog

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// To avoid allocating when assigning to an interface{}, context keys often have concrete type struct{}.
type loggerCtxKey struct{}

var logTracerKey = loggerCtxKey{}

type tracerData struct {
	LogTagName string
	Value      interface{}
}

type Tracer struct {
	tag    string
	values map[string]tracerData // FieldName => Value
}

type OptionsTracer func(tracer *Tracer) error

func WithTag(tag string) OptionsTracer {
	return func(tracer *Tracer) error {
		tracer.tag = tag
		return nil
	}
}

func NewTracer(v interface{}, opts ...OptionsTracer) (*Tracer, error) {
	tracer := &Tracer{
		tag:    "tracer",
		values: make(map[string]tracerData),
	}

	for _, o := range opts {
		err := o(tracer)
		if err != nil {
			return nil, err
		}
	}

	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Struct:
	default:
		return nil, fmt.Errorf("only valid for struct or pointer to struct, current type %s", rv.Kind())
	}

	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem() // dereference pointer
		rt = rt.Elem()
	}

	// iterate all field
	mapLog := make(map[string]struct{})
	for i := 0; i < rv.NumField(); i++ {
		fieldKey := rt.Field(i)
		field := rv.Field(i)

		switch field.Kind() {
		case reflect.Bool: // boolean
		case reflect.String: // string
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // unsigned number
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // number
		case reflect.Float32, reflect.Float64: // decimal

		default:
			return nil,
				fmt.Errorf("tracer struct only valid for simple field's type, not support type %s on field %s",
					field.Kind(), fieldKey.Name)
		}

		fieldName := fieldKey.Name
		logKeyName := getLogKey(fieldKey, tracer.tag)

		// if log tag name inside `tracer:"key"` is duplicate, then error
		if _, ok := tracer.values[logKeyName]; ok {
			return nil, fmt.Errorf("duplicate field name '%s'", logKeyName)
		}

		mapLog[logKeyName] = struct{}{}
		tracer.values[fieldName] = tracerData{
			LogTagName: logKeyName,
			Value:      field.Interface(),
		}
	}

	return tracer, nil
}

func UnmarshalTracer(tracer *Tracer, v interface{}) error {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("cannot pass nil or non-ptr value")
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem() // dereference pointer
		rt = rt.Elem()
	}

	for i := 0; i < rv.NumField(); i++ {
		fieldKey := rt.Field(i)
		field := rv.Field(i)

		fieldName := fieldKey.Name
		value, ok := tracer.values[fieldName] // get value based on Field key name
		if !ok {
			continue
		}

		actualValue := value.Value
		actualValueReflect := reflect.ValueOf(actualValue)
		if !field.CanSet() {
			return fmt.Errorf("target field %s with type %s cannot be set", fieldName, field.Kind())
		}

		if actualValueReflect.Kind() != field.Kind() {
			return fmt.Errorf("target field %s with type %s does not have match source type %s",
				fieldName, field.Kind(), actualValueReflect.Kind())
		}

		field.Set(actualValueReflect)
	}

	return nil
}

func getLogKey(sf reflect.StructField, tagKey string) string {
	tagKey = strings.TrimSpace(tagKey)
	tagValue, ok := sf.Tag.Lookup(tagKey)
	if !ok {
		return sf.Name
	}

	return strings.Split(tagValue, ",")[0]
}

// Inject Tracer object into context.
// As Go doc said: https://golang.org/pkg/context/#WithValue
// Use context Values only for request-scoped data that transits processes and APIs,
// not for passing optional parameters to functions.
// https://blog.golang.org/context
func Inject(ctx context.Context, stuff *Tracer) context.Context {
	return context.WithValue(ctx, logTracerKey, stuff)
}

// Extract get Tracer information from context
func Extract(ctx context.Context) (*Tracer, bool) {
	stuff, ok := ctx.Value(logTracerKey).(*Tracer)
	if !ok {
		return &Tracer{}, false
	}

	return stuff, ok
}

// MustExtract will extract Tracer without false condition.
// When Tracer is not exist, it will return empty Tracer instead of error.
func MustExtract(ctx context.Context) *Tracer {
	stuff, _ := Extract(ctx)
	return stuff
}
