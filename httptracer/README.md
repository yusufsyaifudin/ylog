# Trace HTTP

## Tracing HTTP Client


You can wrap `*http.Client` using log transport.

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"github.com/yusufsyaifudin/ylog/httptracer"
)

type Library struct {
	HTTP *http.Client
}

func (l *Library) callExampleCom() (out []byte, err error) {
	resp, err := l.HTTP.Get("https://example.com")
	return io.ReadAll(resp.Body)
}

func New(client *http.Client) *Library {
	return &Library{
		HTTP: client,
	}
}

// calls will look like this

func main() {
	httpClient := &http.Client{
		Transport: httptracer.HTTPClient(nil),
	}

	lib := New(httpClient)
	out, err := lib.callExampleCom()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

```