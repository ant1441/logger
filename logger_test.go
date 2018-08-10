package logger

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

var (
	myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})
	myHandlerWithError = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
	})
)

func TestNoConfig(t *testing.T) {
	l := New()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/should/be/stdout/", nil)
	req.RemoteAddr = "111.222.333.444"
	l.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")
}

func TestDefaultConfig(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger: logger,
	})

	res := httptest.NewRecorder()
	url := "/foo/wow?q=search-term&print=1#comments"
	req, _ := http.NewRequest("GET", url, nil)
	req.RequestURI = url
	l.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_uri=\"%s\"", url))
}

func TestDefaultConfigPostError(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger: logger,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/foo", nil)
	l.Handler(myHandlerWithError).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusBadGateway)
	expect(t, strings.TrimSpace(res.Body.String()), strings.TrimSpace(http.StatusText(http.StatusBadGateway)))

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusBadGateway))
	expectContainsTrue(t, buf.String(), "http_method=POST")
}

func TestResponseSize(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger: logger,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	l.Handler(myHandler).ServeHTTP(res, req)

	// Result of myHandler should be three bytes.
	expectContainsTrue(t, buf.String(), "http_size=3")
}

func TestCustomMessage(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Message: "some message",
		Logger:  logger,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	l.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), "msg=\"some message\"")
}

func TestCustomFields(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:       logger,
		CustomFields: logrus.Fields{"foo": "bar"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsFalse(t, buf.String(), "foo=\"bar\"")
}

func TestDefaultRemoteAddress(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger: logger,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.RemoteAddr = "8.8.4.4"
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_addr=%s", req.RemoteAddr))
}

func TestDefaultRemoteAddressWithXForwardFor(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:               logger,
		RemoteAddressHeaders: []string{"X-Forwarded-Proto"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.RemoteAddr = "8.8.4.4"
	req.Header.Add("X-Forwarded-Proto", "12.34.56.78")
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), "http_addr=12.34.56.78")
	expectContainsFalse(t, buf.String(), fmt.Sprintf("http_addr=%s", req.RemoteAddr))
}

func TestDefaultRemoteAddressWithXForwardForFallback(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:               logger,
		RemoteAddressHeaders: []string{"X-Forwarded-Proto"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.RemoteAddr = "8.8.4.4"
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_addr=%s", req.RemoteAddr))
}

func TestDefaultRemoteAddressMultiples(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:               logger,
		RemoteAddressHeaders: []string{"X-Real-IP", "X-Forwarded-Proto"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.RemoteAddr = "8.8.4.4"
	req.Header.Add("X-Forwarded-Proto", "12.34.56.78")
	req.Header.Add("X-Real-IP", "98.76.54.32")
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), "http_addr=98.76.54.32")
	expectContainsFalse(t, buf.String(), "http_addr=12.34.56.78")
	expectContainsFalse(t, buf.String(), fmt.Sprintf("http_addr=%s", req.RemoteAddr))
}

func TestDefaultRemoteAddressMultiplesFallback(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:               logger,
		RemoteAddressHeaders: []string{"X-Real-IP", "X-Forwarded-Proto"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.RemoteAddr = "8.8.4.4"
	req.Header.Add("X-Forwarded-Proto", "12.34.56.78")
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsFalse(t, buf.String(), "http_addr=98.76.54.32")
	expectContainsTrue(t, buf.String(), "http_addr=12.34.56.78")
	expectContainsFalse(t, buf.String(), fmt.Sprintf("http_addr=%s", req.RemoteAddr))
}

func TestIgnoreMultipleConfigs(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	opt1 := Options{Logger: logger}
	opt2 := Options{}

	l := New(opt1, opt2)

	res := httptest.NewRecorder()
	url := "/should/output/to/buf/only/"
	req, _ := http.NewRequest("GET", url, nil)
	req.RequestURI = url
	l.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_uri=%s", url))
}

func TestIgnoredURIsNoMatch(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:             logger,
		IgnoredRequestURIs: []string{"/favicon.ico"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	l.Handler(myHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), fmt.Sprintf("http_status=%d", http.StatusOK))
	expectContainsTrue(t, buf.String(), "http_method=GET")
}

func TestIgnoredURIsMatchig(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := logrus.New()
	logger.SetOutput(buf)

	l := New(Options{
		Logger:             logger,
		IgnoredRequestURIs: []string{"/favicon.ico", "/foo"},
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.RequestURI = "/foo"
	l.Handler(myHandler).ServeHTTP(res, req)

	expect(t, buf.String(), "")
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func expectContainsTrue(t *testing.T, a, b string) {
	if !strings.Contains(a, b) {
		t.Errorf("Expected [%s] to contain [%s]", a, b)
	}
}

func expectContainsFalse(t *testing.T, a, b string) {
	if strings.Contains(a, b) {
		t.Errorf("Expected [%s] to contain [%s]", a, b)
	}
}
