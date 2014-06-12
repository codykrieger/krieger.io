package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "path"

  "github.com/gocraft/web"
)

type context struct {}
type gfxContext struct {
  *context
}

func main() {
  workingDirectory, _ := os.Getwd()

  log.Printf("Working directory: %s", workingDirectory)

  router := web.New(context{}).
    Middleware(web.LoggerMiddleware).
    Middleware(web.StaticMiddleware(path.Join(workingDirectory, "_site"))).
    Get("/downloads/:file:gfxCardStatus-(.+)", gfxDownloadsHandler)

  gfxRouter := router.Subrouter(gfxContext{}, "/gfxCardStatus")

  // gocraft/web doesn't support NotFound handlers on anything but the root
  // router...and it also doesn't support optional path segments...so we have
  // to do some really dumb shit here to make nested paths redirect to gfx.io.
  gfxRouter.Get("/appcast.xml", gfxAppcastHandler).
    Get("/:whatever", gfxHandler).
    Get("/:whatever/:whateveragain", gfxHandler).
    Get("/:whatever/:whateveragain/:whateverathirdtime", gfxHandler)

  port := os.Getenv("PORT")
  if len(port) == 0 {
    port = "3000"
  }

  log.Printf("Listening on http://0.0.0.0:%s...", port)

  log.Fatal(http.ListenAndServe(":"+port, router))
}

func gfxDownloadsHandler(w web.ResponseWriter, req *web.Request) {
  scheme := req.URL.Scheme
  file := req.PathParams["file"]
  http.Redirect(w, req.Request, scheme+"://gfx.io/downloads/"+file, http.StatusMovedPermanently)
}

func gfxAppcastHandler(w web.ResponseWriter, req *web.Request) {
  scheme := req.URL.Scheme
  http.Redirect(w, req.Request, scheme+"://gfx.io/appcast.xml", http.StatusMovedPermanently)
}

func gfxHandler(w web.ResponseWriter, req *web.Request) {
  // scheme := req.Request.URL.Scheme
  fmt.Fprintf(w, "yup: %s", req.Request.URL.String())
  // http.Redirect(w, req.Request, scheme+"://gfx.io", http.StatusMovedPermanently)
}
