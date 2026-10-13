package web

import (
	"net"
	"net/http"
	"strings"
)

// checkHost protexe fronte a DNS rebinding. A web non ten contrasinal, así que
// unha páxina maliciosa que fixese resolver o seu dominio a este equipo podería
// ler e cambiar os datos desde o navegador. Só se aceptan peticións dirixidas a
// unha IP (o propio equipo ou a súa IP na rede local, como desde o móbil), a
// "localhost" ou ao nome deste equipo.
func (s *Server) checkHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowedHost(r.Host, s.hostname) {
			if isAPIPath(r.URL.Path) {
				writeError(w, http.StatusForbidden, "enderezo non permitido: usa a IP do equipo ou localhost")
				return
			}
			http.Error(w, "enderezo non permitido: usa a IP do equipo ou localhost", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allowedHost indica se a cabeceira Host (con ou sen porto) é unha IP,
// "localhost" ou o nome do equipo (tamén co sufixo mDNS ".local").
func allowedHost(hostport, machine string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.TrimSuffix(strings.Trim(host, "[]"), ".")
	if i := strings.IndexByte(host, '%'); i >= 0 {
		host = host[:i] // zona dunha IPv6 local (fe80::1%eth0)
	}
	if net.ParseIP(host) != nil {
		return true
	}
	host = strings.ToLower(host)
	if host == "localhost" {
		return true
	}
	machine = strings.ToLower(machine)
	return machine != "" && (host == machine || host == machine+".local")
}
