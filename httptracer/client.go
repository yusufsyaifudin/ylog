package httptracer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/yusufsyaifudin/ylog"
	"go.uber.org/multierr"
)

type RoundTripper struct {
	Base http.RoundTripper
}

var _ http.RoundTripper = (*RoundTripper)(nil)

func NewHTTPClientTracer(roundTripper http.RoundTripper) http.RoundTripper {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	return &RoundTripper{Base: roundTripper}
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t0 := time.Now()

	var (
		ctx  = req.Context() // request context
		resp *http.Response  // final response
		err  error           // final error
	)

	var (
		reqBodyObj interface{}
		reqBody    []byte
		reqBodyErr error
	)

	if req.Body != nil {
		reqBody, reqBodyErr = ioutil.ReadAll(req.Body)
		if reqBodyErr != nil {
			err = multierr.Append(err, fmt.Errorf("error read request body: %w", reqBodyErr))
			reqBody = []byte("")
		}

		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	if _err := json.Unmarshal(reqBody, &reqBodyObj); _err != nil && len(reqBody) > 0 {
		err = multierr.Append(err, fmt.Errorf("request body is invalid json: %w", _err))
	} else {
		reqBody = []byte{} // if valid json, then don't do double log
	}

	var roundTripErr error
	resp, roundTripErr = r.Base.RoundTrip(req.WithContext(ctx))
	if roundTripErr != nil {
		err = multierr.Append(err, fmt.Errorf("error doing actual request: %w", roundTripErr))
	}

	if resp == nil {
		resp = &http.Response{}
	}

	var (
		respObject  interface{}
		respBody    []byte
		respErrBody error
	)

	if resp.Body != nil {
		respBody, respErrBody = ioutil.ReadAll(resp.Body)
		if respErrBody != nil {
			err = multierr.Append(err, fmt.Errorf("error read response body: %w", respErrBody))
			respBody = []byte{}
		}

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))
	}

	if _err := json.Unmarshal(respBody, &respObject); _err != nil && len(respBody) > 0 {
		err = multierr.Append(err, fmt.Errorf("resp body is invalid json: %w", _err))
	} else {
		respBody = []byte{} // if valid json, then don't do double log
	}

	errStr := ""
	if err != nil {
		errStr = err.Error()
		err = nil
	}

	var toSimpleMap = func(h http.Header) map[string]string {
		out := map[string]string{}
		for k, v := range h {
			out[k] = strings.Join(v, " ")
		}

		return out
	}

	// log outgoing request
	ylog.Access(ctx, ylog.AccessLogData{
		Path: req.URL.String(),
		Request: ylog.HTTPData{
			Header:     toSimpleMap(req.Header),
			DataObject: reqBodyObj,
			DataString: string(reqBody),
		},
		Response: ylog.HTTPData{
			Header:     toSimpleMap(resp.Header),
			DataObject: respObject,
			DataString: string(respBody),
		},
		Error:       errStr,
		ElapsedTime: time.Since(t0).Milliseconds(),
	})

	// use only round-tripper error to return.
	// error from marshalling json or read body will ignored
	return resp, roundTripErr
}
