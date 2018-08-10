# Logger-logrus [![GoDoc](https://godoc.org/github.com/ant1441/logger-logrus?status.svg)](http://godoc.org/github.com/ant1441/logger-logrus) [![Build Status](https://travis-ci.org/ant1441/logger-logrus.svg)](https://travis-ci.org/ant1441/logger-logrus)

Logger-logrus is an HTTP middleware for Go that logs web requests to an logrus.Logger (the default being `logrus.StandardLogger()`).
It's a standard net/http [Handler](http://golang.org/pkg/net/http/#Handler), and can be used with many frameworks or directly with Go's net/http package.

This fork of [unrolled/logger](https://github.com/unrolled/logger) was build to support [https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus)

## Usage

~~~ go
// main.go
package main

import (
    "log"
    "net/http"

    "github.com/ant1441/logger-logrus"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello world"))
})

func main() {
    loggerWithConfigMiddleware := logger.New(logger.Options{
        RemoteAddressHeaders: []string{"X-Forwarded-For"},
    })

    // loggerWithDefaults := logger.New()

    app := loggerWithConfigMiddleware.Handler(myHandler)
    http.ListenAndServe("0.0.0.0:3000", app)
}
~~~

A simple GET request to "/info/" will output:
~~~ bash
INFO[0013] Request received                              http_addr="127.0.0.1:41634" http_duration="4.511µs" http_method=GET http_proto=HTTP/1.1 http_size=11 http_status=200 http_uri=/info
~~~

Be sure to use the Logger middleware as the very first handler in the chain. This will ensure that your subsequent handlers (like [Recovery](http://github.com/unrolled/recovery)) will always be logged.

### Available Options
Logger comes with a variety of configuration options (Note: these are not the default option values. See the defaults below.):

~~~ go
// ...
l := logger.New(logger.Options{        
    Message: "Request received", // Message is the outputted log message, default is "Request received"
    CustomFields logrus.Fields, // CustomFields allows passing of custom logging fields, default is empty
    RemoteAddressHeaders: []string{"X-Forwarded-For"}, // RemoteAddressHeaders is a list of header keys that Logger will look at to determine the proper remote address. Useful when using a proxy like Nginx: `[]string{"X-Forwarded-For"}`. Default is an empty slice, and thus will use `reqeust.RemoteAddr`.
    Logger: os.Stdout, // Logger is the logrus.Logger used. Default is logrus.StandardLogger() is used
    IgnoredRequestURIs: []string{"/favicon.ico"}, // IgnoredRequestURIs is a list of path values we do not want logged out. Exact match only!
})
// ...
~~~

### Default Options
These are the preset options for Logger:

~~~ go
l := logger.New()

// Is the same as the default configuration options:

l := logger.New(logger.Options{        
    Message: "Request received",
    RemoteAddressHeaders: []string{},
    Out: logrus.StandardLogger(),
    IgnoredRequestURIs: []string{},
})
~~~

### Capturing the proper remote address
If your app is behind a load balancer or proxy, the default `Request.RemoteAddr` will likely be wrong.
To ensure you're logging the correct IP address, you can set the `RemoteAddressHeaders` option to a list of header names you'd like to use. Logger will iterate over the slice and use the first header value it finds.
If it finds none, it will default to the `Request.RemoteAddr`.

~~~ go
package main

import (
    "log"
    "net/http"

    "github.com/ant1441/logger-logrus"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello world"))
})

func main() {
    loggerWithConfigMiddleware := logger.New(logger.Options{
        RemoteAddressHeaders: []string{"X-Real-IP", "X-Forwarded-For"},
    })

    app := loggerMiddleware.Handler(myHandler)
    http.ListenAndServe("0.0.0.0:3000", app)
}
~~~
