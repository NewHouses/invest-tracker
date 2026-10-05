package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

func DistFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return distFS
	}
	return sub
}

func (s *Server) staticHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "método non permitido", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "." || name == "" {
			serveStaticFile(w, r, s.static, "index.html", false)
			return
		}

		info, err := fs.Stat(s.static, name)
		if err == nil {
			if info.IsDir() {
				http.NotFound(w, r)
				return
			}
			serveStaticFile(w, r, s.static, name, strings.HasPrefix(name, "assets/"))
			return
		}

		// Un ficheiro (con extensión) que non existe é un 404 real: devolver
		// index.html rompería a carga de módulos JS/CSS con cachés antigas.
		if strings.HasPrefix(name, "assets/") || path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		serveStaticFile(w, r, s.static, "index.html", false)
	})
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, fileSystem fs.FS, name string, immutable bool) {
	if immutable {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeFileFS(w, r, fileSystem, name)
}
