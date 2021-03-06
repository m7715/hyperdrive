package hyperdrive

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

// DefaultMiddlewareChain wraps the given http.Handler in the following chain
// of middleware: CorsMiddleware, FrameOptionsMiddleware,
// ContentTypeOptionsMiddleware, CompressionMiddleware, LoggingMiddleware,
// RecoveryMiddleware.
func (api *API) DefaultMiddlewareChain(h http.Handler) http.Handler {
	return api.CorsMiddleware(api.FrameOptionsMiddleware(api.ContentTypeOptionsMiddleware(api.CompressionMiddleware(api.LoggingMiddleware(api.RecoveryMiddleware(h))))))
}

// LoggingMiddleware wraps the given http.Handler and outputs requests in Apache-style
// Combined Log format. All logging is done to STDOUT only.
func (api *API) LoggingMiddleware(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}

// RecoveryMiddleware wraps the given http.Handler and recovers from panics. It wil log
// the stacktrace if HYPERDRIVE_ENVIRONMENT env var is not set to "production".
func (api *API) RecoveryMiddleware(h http.Handler) http.Handler {
	opt := handlers.PrintRecoveryStack(conf.Env != "production")
	return handlers.RecoveryHandler(opt)(h)
}

// CompressionMiddleware wraps the given http.Handler and returns a gzipped response if
// the client requests it with the Accept-Encoding header. The compression level is set
// to to 1, by default. You can configure this though the
// GZIP_LEVEL environment variable, and set it to an integer between -2 and 9.
//
// Following zlib, levels range from 1 (Best Speed) to 9 (Best Compression); higher
// levels typically run slower but compress more.
//
// -1 is the Default Compression level, and is also used if an invalid value is
// configured via GZIP_LEVEL.
//
// 0 attempts no compression, and only adds the necessary DEFLATE framing.
//
// -2 disables Lempel-Ziv match searching and only performs Huffman entropy
// encoding. This is useful when compressing data that has already been compressed
// with an LZ style algorithm, such as Snappy or LZ4.
//
// More info can be found in the docs for the compress/flate package:
// https://golang.org/pkg/compress/flate/
func (api *API) CompressionMiddleware(h http.Handler) http.Handler {
	return handlers.CompressHandlerLevel(h, conf.GzipLevel)
}

// MethodOverrideMiddleware allows clients who can not perform native PUT, PATCH,
// or DELETE requests to specify the HTTP method in the X-HTTP-Method-Override
// header. The header name is case sensitive.
func (api *API) MethodOverrideMiddleware(h http.Handler) http.Handler {
	return handlers.HTTPMethodOverrideHandler(h)
}

// CorsMiddleware allows cross-origin HTTP requests to your API. The middleware is enabled
// by default, and can be configured via the following environment variables:
//
// - CORS_ENABLED (bool)
// - CORS_ORIGINS (string)
// - CORS_HEADERS (string)
// - CORS_CREDENTIALS (bool)
func (api *API) CorsMiddleware(h http.Handler) http.Handler {
	if conf.CorsEnabled != true {
		return h
	}
	defaultHeaders := []string{"Content-Type", "X-Content-Type-Options"}
	headers := handlers.AllowedHeaders(append(defaultHeaders, strings.Split(conf.CorsHeaders, ",")...))
	origins := handlers.AllowedOrigins(strings.Split(conf.CorsOrigins, ","))
	if conf.CorsCredentials == true {
		handlers.AllowCredentials()
	}
	return handlers.CORS(headers, origins)(h)
}

// ContentTypeOptionsMiddleware adds X-Content-Type-Options header set to nosniff to every response.
func (api *API) ContentTypeOptionsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		h.ServeHTTP(rw, r)
	})
}

// FrameOptionsMiddleware adds X-Frame-Options header set to nosniff to every response.
func (api *API) FrameOptionsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("X-Frame-Options", "DENY")
		h.ServeHTTP(rw, r)
	})
}
