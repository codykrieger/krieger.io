package main

import (
	"fmt"
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

var staticMux = http.NewServeMux()

func main() {
	workingDirectory, _ := os.Getwd()

	log.Printf("Working directory: %s", workingDirectory)

	staticMux.Handle("/", http.FileServer(http.Dir(path.Join(workingDirectory, "_site"))))
	http.HandleFunc("/", rootHandler)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	log.Printf("Listening on http://0.0.0.0:%s...", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()

	host := req.Host
	path := req.RequestURI
	scheme := req.Header.Get("X-Forwarded-Proto")
	if len(scheme) == 0 {
		scheme = "http"
	}

	loggingWriter := &loggingResponseWriter{w, 0}

	if strings.HasPrefix(host, "www.") {
		nakedHost := strings.TrimPrefix(host, "www.")
		http.Redirect(loggingWriter, req, scheme+"://"+nakedHost+path, http.StatusMovedPermanently)
	} else if host == "codykrieger.com" {
		http.Redirect(loggingWriter, req, scheme+"://krieger.io"+path, http.StatusMovedPermanently)
	} else if path == "/gfxCardStatus/appcast.xml" {
		http.Redirect(loggingWriter, req, scheme+"://gfx.io/appcast.xml", http.StatusMovedPermanently)
	} else if strings.HasPrefix(path, "/gfxCardStatus") {
		http.Redirect(loggingWriter, req, scheme+"://gfx.io", http.StatusMovedPermanently)
	} else if strings.HasPrefix(path, "/downloads/gfxCardStatus-") {
		http.Redirect(loggingWriter, req, scheme+"://gfx.io"+path, http.StatusMovedPermanently)
	} else {
		staticMux.ServeHTTP(loggingWriter, req)
	}

	duration := time.Since(startTime).Nanoseconds()

	var units string
	switch {
	case duration > 2000000:
		units = "ms"
		duration /= 1000000
	case duration > 1000:
		units = "μs"
		duration /= 1000
	default:
		units = "ns"
	}

	fmt.Printf("[%d %s] %d '%s'\n", duration, units, loggingWriter.status, path)
}
