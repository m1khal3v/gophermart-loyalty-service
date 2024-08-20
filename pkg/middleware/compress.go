package middleware

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
)

type encoderPool struct {
	order []string
	pool  map[string]*sync.Pool
}

type compressedResponseWriter struct {
	http.ResponseWriter
	encoder        io.Writer
	writer         io.Writer
	encoding       string
	supportedTypes []string
	wroteHeader    bool
}

func newEncoderPool(level uint8) *encoderPool {
	return &encoderPool{
		order: []string{"gzip", "deflate"},
		pool: map[string]*sync.Pool{
			"gzip": {
				New: func() any {
					writer, err := gzip.NewWriterLevel(io.Discard, int(level))
					if err != nil {
						return nil
					}

					return writer
				},
			},
			"deflate": {
				New: func() any {
					writer, err := flate.NewWriter(io.Discard, int(level))
					if err != nil {
						return nil
					}

					return writer
				},
			},
		},
	}
}

func getDefaultContentTypes() []string {
	return []string{
		"text/html",
		"text/css",
		"text/plain",
		"text/javascript",
		"application/javascript",
		"application/x-javascript",
		"application/json",
		"application/atom+xml",
		"application/rss+xml",
		"image/svg+xml",
	}
}

func Compress(level uint8, types ...string) func(next http.Handler) http.Handler {
	if len(types) == 0 {
		types = getDefaultContentTypes()
	}
	encoderPool := newEncoderPool(level)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			encoder, encoding, restore := encoderPool.getEncoder(request.Header, writer)
			if encoder == nil {
				next.ServeHTTP(writer, request)
				return
			}

			compressedWriter := &compressedResponseWriter{
				ResponseWriter: writer,
				encoder:        encoder,
				encoding:       encoding,
				supportedTypes: types,
			}

			defer compressedWriter.Close()
			defer restore()

			next.ServeHTTP(compressedWriter, request)
		})
	}
}

type resettableWriter interface {
	io.Writer
	Reset(w io.Writer)
}

func (encoderPool encoderPool) getEncoder(header http.Header, writer http.ResponseWriter) (io.Writer, string, func()) {
	acceptEncoding := header.Get("Accept-Encoding")
	acceptedEncodings := strings.Split(strings.ToLower(acceptEncoding), ",")
	for _, encoding := range encoderPool.order {
		for _, acceptedEncoding := range acceptedEncodings {
			if strings.HasPrefix(strings.TrimLeft(acceptedEncoding, " "), encoding) {
				pool := encoderPool.pool[encoding]
				encoder := pool.Get().(resettableWriter)
				restore := func() {
					pool.Put(encoder)
				}
				encoder.Reset(writer)

				return encoder, encoding, restore
			}
		}
	}

	return nil, "", nil
}

func (writer *compressedResponseWriter) WriteHeader(code int) {
	if writer.wroteHeader {
		writer.ResponseWriter.WriteHeader(code)
		return
	}

	defer writer.ResponseWriter.WriteHeader(code)
	writer.wroteHeader = true
	writer.writer = writer.ResponseWriter

	if writer.Header().Get("Content-Encoding") != "" {
		return
	}

	contentType := writer.Header().Get("Content-Type")
	contentType, _, _ = strings.Cut(contentType, ";")
	if !slices.Contains(writer.supportedTypes, contentType) {
		return
	}

	writer.writer = writer.encoder
	writer.Header().Set("Content-Encoding", writer.encoding)
	writer.Header().Add("Vary", "Accept-Encoding")
	writer.Header().Del("Content-Length")
}

func (writer *compressedResponseWriter) Write(p []byte) (int, error) {
	if !writer.wroteHeader {
		writer.WriteHeader(http.StatusOK)
	}

	return writer.writer.Write(p)
}

type compressFlusher interface {
	Flush() error
}

func (writer *compressedResponseWriter) Flush() {
	if flusher, ok := writer.writer.(compressFlusher); ok {
		_ = flusher.Flush()
	}

	if flusher, ok := writer.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (writer *compressedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := writer.writer.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("chi/middleware/compress: http.Hijacker is unavailable on the writer")
}

func (writer *compressedResponseWriter) Push(target string, opts *http.PushOptions) error {
	if ps, ok := writer.writer.(http.Pusher); ok {
		return ps.Push(target, opts)
	}
	return errors.New("chi/middleware/compress: http.Pusher is unavailable on the writer")
}

func (writer *compressedResponseWriter) Close() error {
	if closer, ok := writer.writer.(io.WriteCloser); ok {
		return closer.Close()
	}
	return errors.New("chi/middleware/compress: io.WriteCloser is unavailable on the writer")
}

func (writer *compressedResponseWriter) Unwrap() http.ResponseWriter {
	return writer.ResponseWriter
}
