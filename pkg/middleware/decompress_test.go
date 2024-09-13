package middleware

import (
	"bytes"
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

func TestDecompress(t *testing.T) {
	tests := []struct {
		name            string
		contentEncoding string
		wantStatusCode  int
	}{
		{
			name:            "no encoding",
			contentEncoding: "",
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "gzip encoding",
			contentEncoding: "gzip",
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "deflate encoding",
			contentEncoding: "deflate",
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "unknown encoding",
			contentEncoding: "lz4",
			wantStatusCode:  http.StatusUnsupportedMediaType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(Decompress())
			router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
				bytes, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatal(err)
				}

				writer.Write(bytes)
			})
			httpServer := httptest.NewServer(router)
			defer httpServer.Close()

			request, err := http.NewRequest(http.MethodGet, httpServer.URL+"/", getBody(t, tt.contentEncoding))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Content-Encoding", tt.contentEncoding)

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			encodings := strings.Split(response.Header.Get("Accept-Encoding"), ", ")
			assert.Contains(t, encodings, "gzip")
			assert.Contains(t, encodings, "deflate")
			assert.Equal(t, tt.wantStatusCode, response.StatusCode)

			if tt.wantStatusCode == http.StatusOK {
				acceptEncodings := response.Header.Get("Accept-Encoding")
				assert.Contains(t, acceptEncodings, "gzip")
				assert.Contains(t, acceptEncodings, "deflate")
				bytes, err := io.ReadAll(response.Body)
				if err != nil {
					return
				}
				assert.Equal(t, "Hello World!", string(bytes))
			}
		})
	}
}

func getBody(t *testing.T, encoding string) io.Reader {
	t.Helper()
	var buffer io.Writer = bytes.NewBuffer(nil)
	writer := buffer

	level := rand.IntN(3) + 3
	switch encoding {
	case "gzip":
		var err error
		writer, err = gzip.NewWriterLevel(buffer, level)
		if err != nil {
			t.Fatal(err)
		}
	case "deflate":
		var err error
		writer, err = flate.NewWriter(buffer, level)
		if err != nil {
			t.Fatal(err)
		}
	}

	if _, err := writer.Write([]byte("Hello World!")); err != nil {
		t.Fatal(err)
	}

	if writer, ok := writer.(io.Closer); ok {
		writer.Close()
	}

	return bytes.NewReader(buffer.(*bytes.Buffer).Bytes())
}
