package adminui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// files contains the production Svelte build.
//
//go:embed dist/*
var files embed.FS

type Handler struct {
	fileServer http.Handler
	files      fs.FS
}

func NewHandler() (*Handler, error) {
	subdirectory, err := fs.Sub(files, "dist")
	if err != nil {
		return nil, err
	}

	return &Handler{
		fileServer: http.FileServer(http.FS(subdirectory)),
		files:      subdirectory,
	}, nil
}

func (h *Handler) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	requestPath := strings.TrimPrefix(request.URL.Path, "/admin")
	requestPath = strings.TrimPrefix(requestPath, "/")

	serveIndex := requestPath == ""

	if !serveIndex {
		cleanPath := path.Clean(requestPath)

		fileInfo, err := fs.Stat(h.files, cleanPath)
		if err != nil || fileInfo.IsDir() {
			// Unknown paths are handled by the Svelte entry point.
			serveIndex = true
		} else {
			requestPath = cleanPath
		}
	}

	clone := request.Clone(request.Context())

	if serveIndex {
		// FileServer serves index.html for "/", without triggering its
		// special index.html redirect.
		clone.URL.Path = "/"
		writer.Header().Set("Cache-Control", "no-store")
	} else {
		clone.URL.Path = "/" + requestPath
		writer.Header().Set(
			"Cache-Control",
			"public, max-age=31536000, immutable",
		)
	}

	h.fileServer.ServeHTTP(writer, clone)
}
