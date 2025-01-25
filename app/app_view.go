package app

import (
	"github.com/neurlang/goruut/helpers/log"
)
import . "github.com/martinarisk/di/dependency_injection"

import "github.com/neurlang/goruut/views"
import (
	"embed"
	"net/http"
	"strings"
)

const defaultFrontendVersion = "v0/"

// Views represents application views.
type Views struct {
	FS *embed.FS
}

// NewAppViews creates new instances of application views.
func (app *App) NewAppViews(_ *DependencyInjection) *Views {
	return &Views{&views.Data}
}

// ServeHTTP serves the HTTP request using the specified writer and request.
func (av *Views) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	uri := request.URL.RequestURI()

	for strings.HasPrefix(uri, "/") {
		uri = uri[1:]
	}
	if uri == "" || strings.HasSuffix(uri, "/") {
		uri += "index.html"
	}
	if !strings.HasPrefix(uri, "v") {
		uri = defaultFrontendVersion + uri
	}
	for strings.Contains(uri, "//") {
		uri = strings.Replace(uri, "//", "/", 1)
	}
	log.Error0(av.HandleFile(writer, uri))
}

// ContentType returns the content type for the specified path.
func (av *Views) ContentType(path string) string {
	var contentType = "application/octet-stream"
	html := strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm")
	css := strings.HasSuffix(path, ".css")
	js := strings.HasSuffix(path, ".js")
	svg := strings.HasSuffix(path, ".svg")
	txt := strings.HasSuffix(path, ".txt")
	if html {
		contentType = "text/html"
	}
	if css {
		contentType = "text/css"
	}
	if js {
		contentType = "text/javascript"
	}
	if svg {
		contentType = "image/svg+xml"
	}
	if txt {
		contentType = "text/plain"
	}
	return contentType
}

// HandleFile handles the file at the specified path for HTTP response.
func (av *Views) HandleFile(w http.ResponseWriter, path string) error {
	data, err := av.FS.ReadFile(path)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", av.ContentType(path))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
