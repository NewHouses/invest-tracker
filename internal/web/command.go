package web

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"invest-tracker/internal/store"
)

type Config struct {
	Addr          string
	DBPath        string
	NoOpen        bool
	ResetPassword bool
}

var openBrowser = func(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func Main(args []string) int {
	cfg := Config{}
	fs := flag.NewFlagSet("web", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&cfg.Addr, "addr", "0.0.0.0:8080", "enderezo no que escoitar")
	fs.StringVar(&cfg.DBPath, "db", "./investimentos.db", "ruta da base de datos SQLite")
	fs.BoolVar(&cfg.NoOpen, "no-open", false, "non abrir o navegador automaticamente")
	fs.BoolVar(&cfg.ResetPassword, "reset-password", false, "eliminar o contrasinal web e pechar todas as sesións")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Uso: invest-tracker web [opcións]")
		fmt.Fprintln(fs.Output(), "Opcións:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "non se pode abrir a base de datos:", err)
		return 1
	}
	defer st.Close()

	if cfg.ResetPassword {
		if err := st.DeleteSetting(passwordSettingKey); err != nil {
			fmt.Fprintln(os.Stderr, "non se puido eliminar o contrasinal:", err)
			return 1
		}
		if err := st.DeleteAllSessions(); err != nil {
			fmt.Fprintln(os.Stderr, "non se puideron pechar as sesións:", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "Contrasinal eliminado. Ao arrancar o modo web, créao de novo desde este ordenador (http://localhost:%s).\n", portFromAddr(cfg.Addr))
		return 0
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		if isAddrInUse(err) {
			fmt.Fprintf(os.Stderr, "non se pode iniciar o servidor: o porto %s xa está en uso\n", portFromAddr(cfg.Addr))
		} else {
			fmt.Fprintln(os.Stderr, "non se pode iniciar o servidor:", err)
		}
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := Run(ctx, st, ln, cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "erro no servidor web:", err)
		return 1
	}
	return 0
}

func Run(ctx context.Context, st *store.Store, ln net.Listener, cfg Config, out io.Writer) error {
	srv := New(st, DistFS(), Options{Logger: log.New(out, "", log.LstdFlags)})
	httpSrv := &http.Server{
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	port := listenerPort(ln)
	localURL := "http://localhost:" + port
	fmt.Fprintln(out, "Control de Investimentos — modo web")
	fmt.Fprintf(out, "  Local:      %s\n", localURL)
	for _, url := range networkURLs(ln, port) {
		fmt.Fprintf(out, "  Rede local: %s\n", url)
	}
	fmt.Fprintln(out, "Preme Ctrl+C para parar.")

	if !cfg.NoOpen {
		if err := openBrowser(localURL); err != nil {
			fmt.Fprintln(out, "non se puido abrir o navegador:", err)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		err := httpSrv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

func listenerPort(ln net.Listener) string {
	if addr, ok := ln.Addr().(*net.TCPAddr); ok {
		return strconv.Itoa(addr.Port)
	}
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err == nil {
		return port
	}
	return ""
}

// networkURLs devolve as URLs accesibles desde outros dispositivos: todas as
// IPv4 privadas se se escoita en tódalas interfaces, a propia IP se se
// escoita nunha concreta, e ningunha se só se escoita en loopback.
func networkURLs(ln net.Listener, port string) []string {
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok || addr.IP == nil || addr.IP.IsUnspecified() {
		return lanURLs(port)
	}
	if addr.IP.IsLoopback() {
		return nil
	}
	return []string{"http://" + net.JoinHostPort(addr.IP.String(), port)}
}

func lanURLs(port string) []string {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var urls []string
	for _, iface := range ifs {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, ok := ipFromAddr(addr)
			if !ok || !isPrivateIPv4(ip) {
				continue
			}
			urls = append(urls, "http://"+ip.String()+":"+port)
		}
	}
	return urls
}

func ipFromAddr(addr net.Addr) (net.IP, bool) {
	switch v := addr.(type) {
	case *net.IPNet:
		return v.IP, true
	case *net.IPAddr:
		return v.IP, true
	default:
		return nil, false
	}
}

func isPrivateIPv4(ip net.IP) bool {
	ip4 := ip.To4()
	return ip4 != nil && ip4.IsPrivate()
}

func isAddrInUse(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "address already in use") || strings.Contains(text, "only one usage") || strings.Contains(text, "enderezo xa está en uso")
}

func portFromAddr(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err == nil {
		return port
	}
	return addr
}
