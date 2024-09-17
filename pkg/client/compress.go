package client

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func (client *Client) compressRequestBody(resty *resty.Client, request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	writer := client.gzipPool.Get().(*gzip.Writer)
	defer client.gzipPool.Put(writer)
	writer.Reset(buffer)

	_, err := io.Copy(writer, request.Body)
	if err = errors.Join(err, writer.Close(), request.Body.Close()); err != nil {
		return err
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
	request.ContentLength = int64(buffer.Len())
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return nil
}
