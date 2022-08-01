package httptracer_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yusufsyaifudin/ylog/httptracer"
	"io"
	"net/http"
	"testing"
)

func TestNewHTTPClientTracer(t *testing.T) {
	t.Run("nil http round tripper", func(t *testing.T) {
		roundTripper := httptracer.NewHTTPClientTracer(nil)
		assert.NotNil(t, roundTripper)
	})

	t.Run("ok", func(t *testing.T) {
		roundTripper := httptracer.NewHTTPClientTracer(http.DefaultTransport)
		assert.NotNil(t, roundTripper)
	})
}

type Library struct {
	BaseURL string
	HTTP    *http.Client
}

func (l *Library) callExampleCom() (out []byte, err error) {
	resp, err := l.HTTP.Get(l.BaseURL)
	return io.ReadAll(resp.Body)
}

func TestRoundTripper_RoundTrip(t *testing.T) {
	url := "https://imdb-api.com/en/API/Search/k_12345678/inception%202010"
	roundTripper := httptracer.NewHTTPClientTracer(http.DefaultTransport)

	httpClient := http.DefaultClient
	httpClient.Transport = roundTripper

	lib := &Library{
		BaseURL: url,
		HTTP:    httpClient,
	}

	t.Run("ok", func(t *testing.T) {
		out, err := lib.callExampleCom()
		assert.NotNil(t, out)
		assert.NoError(t, err)

		fmt.Println("OUTPUT from server", string(out))
	})

}
