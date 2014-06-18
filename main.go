package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// var staticMux = http.NewServeMux()

type context struct {
	router *http.Handler
}

func (c *context) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writer := &loggingResponseWriter{w, 0}
	c.router.ServeHTTP(writer, r)
}

func main() {
	workingDirectory, _ := os.Getwd()
	siteDirectory := path.Join(workingDirectory, "_site")

	log.Printf("Working directory: %s", workingDirectory)

	// staticMux.Handle("/", http.FileServer(http.Dir(siteDirectory)))
	// http.HandleFunc("/", rootHandler)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	router := &httprouter.New()
	// router.GET("/downloads/*path", downloadsHandler)
	router.GET("/gfxCardStatus/*path", gfxHandler)
	router.ServeFiles("/*filepath", http.Dir(siteDirectory))

	log.Printf("Listening on http://0.0.0.0:%s...", port)

	c := context{router}
	log.Fatal(http.ListenAndServe(":"+port, c))
}

// func rootHandler(w http.ResponseWriter, req *http.Request) {
//   startTime := time.Now()

//   host := req.Host
//   path := req.RequestURI
//   scheme := req.Header.Get("X-Forwarded-Proto")
//   if len(scheme) == 0 {
//     scheme = "http"
//   }

//   loggingWriter := &loggingResponseWriter{w, 0}

//   if strings.HasPrefix(host, "www.") {
//     nakedHost := strings.TrimPrefix(host, "www.")
//     http.Redirect(loggingWriter, req, scheme+"://"+nakedHost+path, http.StatusMovedPermanently)
//   } else if host == "codykrieger.com" {
//     http.Redirect(loggingWriter, req, scheme+"://krieger.io"+path, http.StatusMovedPermanently)
//   } else if path == "/gfxCardStatus/appcast.xml" {
//     http.Redirect(loggingWriter, req, scheme+"://gfx.io/appcast.xml", http.StatusMovedPermanently)
//   } else if strings.HasPrefix(path, "/gfxCardStatus") {
//     http.Redirect(loggingWriter, req, scheme+"://gfx.io", http.StatusMovedPermanently)
//   } else if strings.HasPrefix(path, "/downloads/gfxCardStatus-") {
//     http.Redirect(loggingWriter, req, scheme+"://gfx.io"+path, http.StatusMovedPermanently)
//   } else {
//     staticMux.ServeHTTP(loggingWriter, req)
//   }

//   duration := time.Since(startTime).Nanoseconds()

//   var units string
//   switch {
//   case duration > 2000000:
//     units = "ms"
//     duration /= 1000000
//   case duration > 1000:
//     units = "μs"
//     duration /= 1000
//   default:
//     units = "ns"
//   }

//   fmt.Printf("[%d %s] %d '%s'\n", duration, units, loggingWriter.status, path)
// }
