package web

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"invest-tracker/internal/store"
)

type Options struct {
	Logger *log.Logger
	Now    func() time.Time
}

type Server struct {
	store        *store.Store
	static       fs.FS
	mux          *http.ServeMux
	log          *log.Logger
	now          func() time.Time
	pbkdf2Iter   int
	loginLimiter *loginLimiter
}

func New(st *store.Store, static fs.FS, opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	s := &Server{
		store:        st,
		static:       static,
		mux:          http.NewServeMux(),
		log:          logger,
		now:          now,
		pbkdf2Iter:   defaultPBKDF2Iter,
		loginLimiter: newLoginLimiter(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	var h http.Handler = s.mux
	h = http.NewCrossOriginProtection().Handler(h)
	h = s.securityHeaders(h)
	h = s.requestLogging(h)
	h = s.recoverPanic(h)
	return h
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Printf("pánico atendendo %s %s: %v", r.Method, r.URL.Path, rec)
				if isAPIPath(r.URL.Path) {
					writeError(w, http.StatusInternalServerError, "erro interno")
					return
				}
				http.Error(w, "erro interno", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := s.now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		s.log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, s.now().Sub(start))
	})
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Frame-Options", "DENY")
		if !isAPIPath(r.URL.Path) {
			h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; font-src 'self' data:; frame-ancestors 'none'")
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) ReadFrom(src io.Reader) (int64, error) {
	if rf, ok := r.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	return io.Copy(r.ResponseWriter, src)
}

func isAPIPath(p string) bool {
	return p == "/api" || strings.HasPrefix(p, "/api/")
}
