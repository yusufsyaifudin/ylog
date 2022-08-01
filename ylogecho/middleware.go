package ylogecho

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/yusufsyaifudin/ylog"
	"go.uber.org/multierr"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func Middleware() echo.MiddlewareFunc {
	var toSimpleMap = func(h http.Header) map[string]string {
		out := map[string]string{}
		for k, v := range h {
			out[k] = strings.Join(v, " ")
		}

		return out
	}

	fn := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t0 := time.Now()

			var (
				err        error
				reqBody    []byte
				reqBodyErr error
				reqBodyObj interface{}
			)

			req := c.Request()
			resp := c.Response()

			// propagate some data
			type ContextData struct {
				RequestID string `tracer:"trace_id"`
				Timestamp string `tracer:"correlation_id"`
			}

			requestID := strings.TrimSpace(req.Header.Get("X-Request-ID"))
			propagateData := ContextData{
				RequestID: requestID,
				Timestamp: time.Now().Format(time.RFC3339Nano),
			}

			tracer, err := ylog.NewTracer(propagateData, ylog.WithTag("tracer"))
			if err != nil {
				err = fmt.Errorf("cannot inject propagation data: %w", err)
				return err
			}

			// add context
			ctx := ylog.Inject(c.Request().Context(), tracer)
			c.SetRequest(c.Request().WithContext(ctx))

			if req.Body != nil {
				reqBody, reqBodyErr = ioutil.ReadAll(req.Body)
				if reqBodyErr != nil {
					err = multierr.Append(err, fmt.Errorf("error read request body: %w", reqBodyErr))
					reqBody = []byte("")
				}

				c.Request().Body = io.NopCloser(bytes.NewReader(reqBody))
			}

			if _err := json.Unmarshal(reqBody, &reqBodyObj); _err == nil {
				reqBody = []byte("")
			}

			// capture response body
			resBody := &bytes.Buffer{}
			mw := io.MultiWriter(resp.Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: resp.Writer}
			resp.Writer = writer
			resp.Header().Set("X-Request-ID", propagateData.RequestID)

			if _err := next(c); _err != nil {
				err = multierr.Append(err, _err)
				//c.Error(err) // doesn't need call this because if we call this, it will duplicate response
			}

			respBody := resBody.Bytes()
			var respObj interface{}
			if _err := json.Unmarshal(respBody, &respObj); _err == nil {
				respBody = []byte("")
			}

			errStr := ""
			if err != nil {
				errStr = err.Error()
			}

			// log outgoing request
			logData := ylog.AccessLogData{
				Path: c.Path(),
				Request: ylog.HTTPData{
					Header:     toSimpleMap(req.Header),
					DataObject: reqBody,
					DataString: string(reqBody),
				},
				Response: ylog.HTTPData{
					Header:     toSimpleMap(resp.Header()),
					DataObject: respObj,
					DataString: string(respBody),
				},
				Error:       errStr,
				ElapsedTime: time.Since(t0).Milliseconds(),
			}

			ylog.Access(ctx, logData)
			return err
		}
	}

	return fn
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

var _ io.Writer = (*bodyDumpResponseWriter)(nil)
var _ http.ResponseWriter = (*bodyDumpResponseWriter)(nil)

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
