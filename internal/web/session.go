package web

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// sessionCookieName é o nome da cookie de sesión do navegador.
const sessionCookieName = "it_session"

// hashToken devolve o SHA-256 en hexadecimal dun token de sesión: na base de
// datos só se garda o hash, nunca o token.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// sessionTokenHash devolve o hash do token da cookie de sesión da petición,
// ou "" se non hai cookie.
func sessionTokenHash(r *http.Request) string {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return ""
	}
	return hashToken(c.Value)
}

// requireSession só deixa pasar as peticións cunha sesión válida e non
// caducada; o resto reciben 401.
func (s *Server) requireSession(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHash := sessionTokenHash(r)
		if tokenHash == "" {
			writeError(w, http.StatusUnauthorized, "autenticación necesaria")
			return
		}
		ok, err := s.store.SessionValid(tokenHash, s.now())
		if err != nil {
			s.internalError(w, r, err)
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "autenticación necesaria")
			return
		}
		h.ServeHTTP(w, r)
	})
}
