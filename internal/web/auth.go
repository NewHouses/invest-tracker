package web

import (
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	rememberDuration = 30 * 24 * time.Hour
	sessionDuration  = 12 * time.Hour
)

type authStatusResponse struct {
	SetupRequired bool `json:"setupRequired"`
	Authenticated bool `json:"authenticated"`
	SetupAllowed  bool `json:"setupAllowed"`
}

type authOKResponse struct {
	Authenticated bool `json:"authenticated"`
}

func (s *Server) registerAuthRoutes() {
	s.handlePublic("GET /api/auth/status", s.authStatus)
	s.handlePublic("POST /api/auth/setup", s.authSetup)
	s.handlePublic("POST /api/auth/login", s.authLogin)
	s.handlePublic("POST /api/auth/logout", s.authLogout)
	s.mux.Handle("POST /api/auth/password", s.requireSession(http.HandlerFunc(s.authChangePassword)))
}

func (s *Server) authStatus(w http.ResponseWriter, r *http.Request) {
	_, hasPassword, err := s.store.GetSetting(passwordSettingKey)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	authenticated := false
	if tokenHash := sessionTokenHash(r); tokenHash != "" {
		authenticated, err = s.store.SessionValid(tokenHash, s.now())
		if err != nil {
			s.internalError(w, r, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, authStatusResponse{SetupRequired: !hasPassword, Authenticated: authenticated, SetupAllowed: isLocalRequest(r)})
}

func (s *Server) authSetup(w http.ResponseWriter, r *http.Request) {
	_, hasPassword, err := s.store.GetSetting(passwordSettingKey)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if hasPassword {
		writeError(w, http.StatusConflict, "o contrasinal xa está configurado")
		return
	}
	if !isLocalRequest(r) {
		writeError(w, http.StatusForbidden, "o contrasinal só se pode crear desde este ordenador")
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if fields := validatePasswordField("password", req.Password); len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	encoded, err := hashPassword(req.Password, s.pbkdf2Iter)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if err := s.store.SetSetting(passwordSettingKey, encoded); err != nil {
		s.internalError(w, r, err)
		return
	}
	if err := s.createSessionCookie(w, rememberDuration, true); err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, authOKResponse{Authenticated: true})
}

func (s *Server) authLogin(w http.ResponseWriter, r *http.Request) {
	encoded, hasPassword, err := s.store.GetSetting(passwordSettingKey)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if !hasPassword {
		writeError(w, http.StatusConflict, "cómpre crear primeiro o contrasinal")
		return
	}
	now := s.now()
	if ok, retry := s.loginLimiter.allow(r, now); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(retry)))
		writeError(w, http.StatusTooManyRequests, "demasiados intentos; téntao de novo máis tarde")
		return
	}
	var req struct {
		Password string `json:"password"`
		Remember bool   `json:"remember"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ok, err := verifyPassword(req.Password, encoded)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if !ok {
		s.loginLimiter.failure(r, now)
		writeError(w, http.StatusUnauthorized, "contrasinal incorrecto")
		return
	}
	s.loginLimiter.success(r)
	if _, err := s.store.DeleteExpiredSessions(now); err != nil {
		s.internalError(w, r, err)
		return
	}
	duration := sessionDuration
	remember := false
	if req.Remember {
		duration = rememberDuration
		remember = true
	}
	if err := s.createSessionCookie(w, duration, remember); err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, authOKResponse{Authenticated: true})
}

func (s *Server) authLogout(w http.ResponseWriter, r *http.Request) {
	if tokenHash := sessionTokenHash(r); tokenHash != "" {
		if err := s.store.DeleteSession(tokenHash); err != nil {
			s.internalError(w, r, err)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) authChangePassword(w http.ResponseWriter, r *http.Request) {
	encoded, hasPassword, err := s.store.GetSetting(passwordSettingKey)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if !hasPassword {
		writeError(w, http.StatusConflict, "cómpre crear primeiro o contrasinal")
		return
	}
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ok, err := verifyPassword(req.CurrentPassword, encoded)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if !ok {
		writeFieldErrors(w, map[string]string{"currentPassword": "o contrasinal actual non é correcto"})
		return
	}
	if fields := validatePasswordField("newPassword", req.NewPassword); len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	newHash, err := hashPassword(req.NewPassword, s.pbkdf2Iter)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if err := s.store.SetSetting(passwordSettingKey, newHash); err != nil {
		s.internalError(w, r, err)
		return
	}
	if err := s.store.DeleteSessionsExcept(sessionTokenHash(r)); err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) createSessionCookie(w http.ResponseWriter, duration time.Duration, persistent bool) error {
	token, err := newSessionToken()
	if err != nil {
		return err
	}
	if err := s.store.CreateSession(hashToken(token), s.now().Add(duration)); err != nil {
		return err
	}
	cookie := &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode}
	if persistent {
		cookie.MaxAge = int(duration / time.Second)
	}
	http.SetCookie(w, cookie)
	return nil
}

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isLocalRequest(r *http.Request) bool {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	remoteIP := net.ParseIP(remoteHost)
	if remoteIP == nil || !remoteIP.IsLoopback() {
		return false
	}
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.TrimPrefix(strings.TrimSuffix(strings.ToLower(host), "]"), "[")
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
