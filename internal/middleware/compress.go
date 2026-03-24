package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type conditionalCompressWriter struct {
	gzipRespWriter
	compress *bool
}

func (ccw conditionalCompressWriter) Write(b []byte) (int, error) {
	//При первом вызове Write проверить выполнение условий для сжатия ответа
	if ccw.compress == nil {
		ct := ccw.rw.Header().Get("Content-Type")
		if strings.Contains(ct, "application/json") ||
			strings.Contains(ct, "text/html") {
			*ccw.compress = true
		} else {
			*ccw.compress = false
		}
	}
	if *ccw.compress {
		return ccw.zw.Write(b)
	} else {
		return ccw.rw.Write(b)
	}
}

type gzipRespWriter struct {
	rw http.ResponseWriter
	zw *gzip.Writer
}

func newGzipRespWriter(rw http.ResponseWriter) gzipRespWriter {
	return gzipRespWriter{
		rw: rw,
		zw: gzip.NewWriter(rw),
	}
}

func (grw gzipRespWriter) Header() http.Header {
	return grw.rw.Header()
}

func (grw gzipRespWriter) WriteHeader(status int) {
	grw.rw.WriteHeader(status)
}

func (grw gzipRespWriter) Write(b []byte) (int, error) {
	return grw.zw.Write(b)
}

func (grw gzipRespWriter) Close() error {
	return grw.zw.Close()
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

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ow := rw

		acceptEncoding := strings.Join(r.Header["Accept-Encoding"], ",")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := conditionalCompressWriter{
				gzipRespWriter: newGzipRespWriter(rw),
			}
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
	}
}
