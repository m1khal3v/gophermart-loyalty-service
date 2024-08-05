package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"strings"
	"sync"
)

type decoderPool struct {
	pool map[string]*sync.Pool
}

func newDecoderPool() *decoderPool {
	return &decoderPool{
		pool: map[string]*sync.Pool{
			"gzip": {
				New: func() any {
					return new(gzip.Reader)
				},
			},
			"deflate": {
				New: func() any {
					return flate.NewReader(bytes.NewReader(nil))
				},
			},
		},
	}
}

func Decompress() func(next http.Handler) http.Handler {
	decoderPool := newDecoderPool()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Accept-Encoding", strings.Join(maps.Keys(decoderPool.pool), ", "))
			encoding := request.Header.Get("Content-Encoding")
			if encoding == "" {
				next.ServeHTTP(writer, request)
				return
			}

			decoder, restore := decoderPool.getDecoder(encoding, request.Body)
			if decoder == nil {
				writer.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			defer decoder.Close()
			defer restore()

			request.Body = decoder
			next.ServeHTTP(writer, request)
		})
	}
}

type gzipResetter interface {
	Reset(r io.Reader) error
}

func (decoderPool decoderPool) getDecoder(encoding string, body io.ReadCloser) (io.ReadCloser, func()) {
	pool, ok := decoderPool.pool[encoding]
	if !ok {
		return nil, nil
	}

	decoder := pool.Get()
	restore := func() {
		pool.Put(decoder)
	}

	switch encoding {
	case "gzip":
		if err := decoder.(gzipResetter).Reset(body); err != nil {
			return nil, nil
		}
	case "deflate":
		if err := decoder.(flate.Resetter).Reset(body, nil); err != nil {
			return nil, nil
		}
	}

	return decoder.(io.ReadCloser), restore
}
