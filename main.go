package main

import (
  "log"
  "net/http"
  "os"
  "path"

  "github.com/gocraft/web"
)

type Context struct {}

func main() {
  workingDirectory, _ := os.Getwd()

  log.Printf("Working directory: %s", workingDirectory)

  router := web.New(Context{}).
    Middleware(web.LoggerMiddleware).
    Middleware(web.StaticMiddleware(path.Join(workingDirectory, "_site")))

  port := os.Getenv("PORT")
  if len(port) == 0 {
    port = "3000"
  }

  log.Printf("Listening on http://0.0.0.0:%s...", port)

  log.Fatal(http.ListenAndServe(":"+port, router))
}
