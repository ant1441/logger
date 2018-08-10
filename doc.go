/*Package logger is an HTTP middleware for Go that logs web requests to a logrus.Logger.

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
      loggerMiddleware := logger.New(logger.Options{
          RemoteAddressHeaders: []string{"X-Forwarded-Proto"},
      })

      // loggerWithDefaults := logger.New()

      app := loggerMiddleware.Handler(myHandler)
      http.ListenAndServe("0.0.0.0:3000", app)
  }

A simple GET request to "/info/" will output:

  INFO[0013] Request received                              http_addr="127.0.0.1:41634" http_duration="4.511µs" http_method=GET http_proto=HTTP/1.1 http_size=11 http_status=200 http_uri=/info
*/
package logger
