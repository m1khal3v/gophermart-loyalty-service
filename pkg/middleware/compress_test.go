package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCompress(t *testing.T) {
	tests := []struct {
		name            string
		types           []string
		contentType     string
		acceptEncodings []string
		wantEncoding    string
	}{
		{
			name:            "no accept encoding",
			types:           []string{"text/plain"},
			contentType:     "text/plain",
			acceptEncodings: []string{},
			wantEncoding:    "",
		},
		{
			name:            "unknown accept encoding",
			types:           []string{"text/plain"},
			contentType:     "text/plain",
			acceptEncodings: []string{"lz4"},
			wantEncoding:    "",
		},
		{
			name:            "content type mismatch",
			types:           []string{"text/plain"},
			contentType:     "text/html",
			acceptEncodings: []string{"gzip"},
			wantEncoding:    "",
		},
		{
			name:            "gzip encoding",
			types:           []string{"text/html"},
			contentType:     "text/html",
			acceptEncodings: []string{"gzip"},
			wantEncoding:    "gzip",
		},
		{
			name:            "deflate encoding",
			types:           []string{"text/html"},
			contentType:     "text/html",
			acceptEncodings: []string{"deflate"},
			wantEncoding:    "deflate",
		},
		{
			name:            "gzip preferred encoding",
			types:           []string{"text/html"},
			contentType:     "text/html",
			acceptEncodings: []string{"deflate", "gzip"},
			wantEncoding:    "gzip",
		},
		{
			name:            "multiple types",
			types:           []string{"text/html", "text/plain"},
			contentType:     "text/plain",
			acceptEncodings: []string{"deflate", "gzip"},
			wantEncoding:    "gzip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(Compress(uint8(rand.UintN(3)+3), tt.types...))
			router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", tt.contentType)
				writer.Write([]byte("Hello World!"))
			})
			httpServer := httptest.NewServer(router)
			defer httpServer.Close()

			request, err := http.NewRequest(http.MethodGet, httpServer.URL+"/", nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Accept-Encoding", strings.Join(tt.acceptEncodings, " , "))

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantEncoding, response.Header.Get("Content-Encoding"))
			assert.Equal(t, "Hello World!", decodeResponseBody(t, response))
		})
	}
}

func decodeResponseBody(t *testing.T, response *http.Response) string {
	t.Helper()
	reader := response.Body
	defer response.Body.Close()

	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			t.Fatal(err)
		}
	case "deflate":
		reader = flate.NewReader(response.Body)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	return string(body)
}
