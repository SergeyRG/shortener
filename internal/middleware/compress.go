package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/SergeyRG/shortener/internal/logging"
	"go.uber.org/zap"
)

type gzipRespWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newGzipRespWriter(rw http.ResponseWriter) *gzipRespWriter {
	return &gzipRespWriter{
		ResponseWriter: rw,
		zw:             nil,
	}
}

func (grw *gzipRespWriter) Write(b []byte) (int, error) {
	if grw.zw != nil {
		return grw.zw.Write(b)
	}

	//При первом вызове Write проверить выполнение условий для сжатия ответа
	ct := grw.Header().Get("Content-Type")

	shouldCompress := strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "text/html")

	if shouldCompress {
		grw.Header().Set("Content-Encoding", "gzip")
		grw.Header().Del("Content-Length")
		grw.zw = gzip.NewWriter(grw.ResponseWriter)

		return grw.zw.Write(b)
	}

	return grw.ResponseWriter.Write(b)
}

func (grw *gzipRespWriter) WriteHeader(statusCode int) {
	ct := grw.Header().Get("Content-Type")
	shouldCompress := strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html")

	if shouldCompress {
		grw.Header().Set("Content-Encoding", "gzip")
		grw.Header().Del("Content-Length")
		grw.zw = gzip.NewWriter(grw.ResponseWriter)
	}
	grw.ResponseWriter.WriteHeader(statusCode)
}

func (grw *gzipRespWriter) Close() error {
	if grw.zw != nil {
		return grw.zw.Close()
	}
	return nil
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ow := rw
		logging.Logger.Debug("получен запрос",
			zap.String("Header", strings.Join(r.Header["Accept-Encoding"], ",")))

		acceptEncoding := strings.Join(r.Header["Accept-Encoding"], ",")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newGzipRespWriter(rw)
			ow = cw
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			gr, err := newGzipReader(r.Body)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = gr
			defer gr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}
