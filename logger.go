package logger

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Options is a struct for specifying configuration parameters for the Logger middleware.
type Options struct {
	// Message is the outputted log message, default is "Request received"
	Message string
	// CustomFields allows passing of custom logging fields
	CustomFields logrus.Fields
	// RemoteAddressHeaders is a list of header keys that Logger will look at to determine the proper remote address. Useful when using a proxy like Nginx: `[]string{"X-Forwarded-Proto"}`. Default is an empty slice, and thus will use `reqeust.RemoteAddr`.
	RemoteAddressHeaders []string
	// Logger is the logrus.Logger used. If not given, logrus.StandardLogger() is used
	Logger *logrus.Logger
	// IgnoredRequestURIs is a list of path values we do not want logged out. Exact match only!
	IgnoredRequestURIs []string
}

// Logger is a HTTP middleware handler that logs a request. Outputted information includes status, method, URL, remote address, size, and the time it took to process the request.
type Logger struct {
	opt Options
}

// New returns a new Logger instance.
func New(opts ...Options) *Logger {
	var o Options
	if len(opts) == 0 {
		o = Options{}
	} else {
		o = opts[0]
	}

	// Determine message.
	if len(o.Message) == 0 {
		o.Message = "Request received"
	}

	// Determine output logger.
	if o.Logger == nil {
		// Default is logrus Standard Logger.
		o.Logger = logrus.StandardLogger()
	}

	return &Logger{
		opt: o,
	}
}

// Handler wraps an HTTP handler and logs the request as necessary.
func (l *Logger) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		crw := newCustomResponseWriter(w)
		next.ServeHTTP(crw, r)

		for _, ignoredURI := range l.opt.IgnoredRequestURIs {
			if ignoredURI == r.RequestURI {
				return
			}
		}

		addr := r.RemoteAddr
		for _, headerKey := range l.opt.RemoteAddressHeaders {
			if val := r.Header.Get(headerKey); len(val) > 0 {
				addr = val
				break
			}
		}

		l.opt.Logger.WithFields(logrus.Fields{
			"http_addr":     addr,
			"http_method":   r.Method,
			"http_uri":      r.RequestURI,
			"http_proto":    r.Proto,
			"http_status":   crw.status,
			"http_size":     crw.size,
			"http_duration": time.Since(start),
		}).WithFields(l.opt.CustomFields).Info(l.opt.Message)
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (c *customResponseWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *customResponseWriter) Write(b []byte) (int, error) {
	size, err := c.ResponseWriter.Write(b)
	c.size += size
	return size, err
}

func (c *customResponseWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (c *customResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := c.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement the Hijacker interface")
}

func newCustomResponseWriter(w http.ResponseWriter) *customResponseWriter {
	// When WriteHeader is not called, it's safe to assume the status will be 200.
	return &customResponseWriter{
		ResponseWriter: w,
		status:         200,
	}
}
