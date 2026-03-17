// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package common

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/andybalholm/brotli/matchfinder"
	"github.com/labstack/echo/v4"
)

type Skipper func(c echo.Context) bool

// BrotliConfig defines the config for Brotli middleware.
type BrotliConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Brotli compression level.
	// Optional. Default value -1.
	Level int `yaml:"level"`

	// Length threshold before brotli compression is applied.
	// Key considerations
	// Small files (under 64 KB): For dynamic content, the overhead of compressing very small files might outweigh the benefits, as Brotli can be slower than gzip in these cases.
	// Large files (over 64 KB): For larger files, Brotli's superior compression ratios provide significant advantages.
	// Performance trade-offs:
	//     Static files: Since static files are pre-compressed, using Brotli is generally a clear win.
	//     Dynamic files: For dynamic content, consider the trade-off between compression speed and the resulting file size. You may choose a lower Brotli setting for speed or a higher one for better compression
	MinLength int
}

type brotliResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteHeader       bool
	wroteBody         bool
	minLength         int
	minLengthExceeded bool
	buffer            *bytes.Buffer
	code              int
}

const (
	brotliScheme = "br"
)

// DefaultBrotliConfig is the default Brotli middleware config.
var DefaultBrotliConfig = BrotliConfig{
	Skipper:   DefaultSkipper,
	Level:     brotli.DefaultCompression,
	MinLength: 0,
}

// Brotli returns a middleware which compresses HTTP response using Brotli compression
// scheme.
func Brotli() echo.MiddlewareFunc {
	return BrotliWithConfig(DefaultBrotliConfig)
}

// BrotliWithConfig return Brotli middleware with config.
// See: `Brotli()`.
func BrotliWithConfig(config BrotliConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultBrotliConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultBrotliConfig.Level
	}
	if config.MinLength < 0 {
		config.MinLength = DefaultBrotliConfig.MinLength
	}

	pool := brotliCompressPool(config)
	bpool := bufferPool()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), brotliScheme) {
				i := pool.Get()
				w, ok := i.(*matchfinder.Writer)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
				}
				rw := res.Writer
				w.Reset(rw)

				buf := bpool.Get().(*bytes.Buffer)
				buf.Reset()

				grw := &brotliResponseWriter{Writer: w, ResponseWriter: rw, minLength: config.MinLength, buffer: buf}
				defer func() {
					// There are different reasons for cases when we have not yet written response to the client and now need to do so.
					// a) handler response had only response code and no response body (ala 404 or redirects etc). Response code need to be written now.
					// b) body is shorter than our minimum length threshold and being buffered currently and needs to be written
					if !grw.wroteBody {
						if res.Header().Get(echo.HeaderContentEncoding) == brotliScheme {
							res.Header().Del(echo.HeaderContentEncoding)
						}
						if grw.wroteHeader {
							rw.WriteHeader(grw.code)
						}
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.Writer = rw
						w.Reset(io.Discard)
					} else if !grw.minLengthExceeded {
						// Write uncompressed response
						res.Writer = rw
						if grw.wroteHeader {
							grw.ResponseWriter.WriteHeader(grw.code)
						}
						grw.buffer.WriteTo(rw)
						w.Reset(io.Discard)
					}
					w.Close()
					bpool.Put(buf)
					pool.Put(w)
				}()
				res.Writer = grw
			}
			return next(c)
		}
	}
}

func (w *brotliResponseWriter) WriteHeader(code int) {
	w.Header().Del(echo.HeaderContentLength) // Issue #444

	w.wroteHeader = true

	// Delay writing of the header until we know if we'll actually compress the response
	w.code = code
}

func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true

	if !w.minLengthExceeded {
		n, err := w.buffer.Write(b)

		if w.buffer.Len() >= w.minLength {
			w.minLengthExceeded = true

			// The minimum length is exceeded, add Content-Encoding header and write the header
			w.Header().Set(echo.HeaderContentEncoding, brotliScheme) // Issue #806
			if w.wroteHeader {
				w.ResponseWriter.WriteHeader(w.code)
			}

			return w.Writer.Write(w.buffer.Bytes())
		}

		return n, err
	}

	return w.Writer.Write(b)
}

func (w *brotliResponseWriter) Flush() {
	if !w.minLengthExceeded {
		// Enforce compression because we will not know how much more data will come
		w.minLengthExceeded = true
		w.Header().Set(echo.HeaderContentEncoding, brotliScheme) // Issue #806
		if w.wroteHeader {
			w.ResponseWriter.WriteHeader(w.code)
		}

		w.Writer.Write(w.buffer.Bytes())
	}

	// w.Writer.(*matchfinder.Writer).Flush()
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *brotliResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *brotliResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *brotliResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func brotliCompressPool(config BrotliConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w := brotli.NewWriterV2(io.Discard, config.Level)
			return w
		},
	}
	// return sync.Pool{
	// 	New: func() interface{} {
	// 		w, err := gzip.NewWriterLevel(io.Discard, config.Level)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		return w
	// 	},
	// }
}

func bufferPool() sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			b := &bytes.Buffer{}
			return b
		},
	}
}

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(echo.Context) bool {
	return false
}
