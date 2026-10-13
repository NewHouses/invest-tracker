// Command macos empaqueta a app de macOS de Control de Investimentos nun .zip
// (dist-installers/ControlInvestimentos-<versión>-macos.zip) con
// "Control de Investimentos.app", que leva os binarios para Apple Silicon e
// Intel. Execútase desde a raíz do repositorio despois de compilar:
//
//	go run ./installers/macos
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const appDir = "Control de Investimentos.app/"

// modTime fixa a data dos ficheiros do zip para que o mesmo contido xere
// sempre o mesmo ficheiro.
var modTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type entry struct {
	name string
	data []byte
	mode fs.FileMode
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run() error {
	version, err := readText("VERSION")
	if err != nil {
		return err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("o ficheiro VERSION está baleiro")
	}
	info, err := readText("installers/macos/Info.plist")
	if err != nil {
		return err
	}
	launcher, err := readText("installers/macos/launcher.sh")
	if err != nil {
		return err
	}
	server, err := readText("installers/macos/servidor.command")
	if err != nil {
		return err
	}
	readme, err := readText("installers/macos/LEME.txt")
	if err != nil {
		return err
	}
	readmeEN, err := readText("installers/macos/README.txt")
	if err != nil {
		return err
	}
	icon, err := os.ReadFile("installers/macos/AppIcon.icns")
	if err != nil {
		return err
	}
	arm64, err := os.ReadFile("invest-tracker-macos-arm64")
	if err != nil {
		return fmt.Errorf("%w (compila primeiro con build.ps1 ou build.sh)", err)
	}
	amd64, err := os.ReadFile("invest-tracker-macos-amd64")
	if err != nil {
		return fmt.Errorf("%w (compila primeiro con build.ps1 ou build.sh)", err)
	}

	entries := []entry{
		{appDir, nil, fs.ModeDir | 0o755},
		{appDir + "Contents/", nil, fs.ModeDir | 0o755},
		{appDir + "Contents/Info.plist", []byte(strings.ReplaceAll(info, "{{VERSION}}", version)), 0o644},
		{appDir + "Contents/PkgInfo", []byte("APPL????"), 0o644},
		{appDir + "Contents/MacOS/", nil, fs.ModeDir | 0o755},
		{appDir + "Contents/MacOS/ControlInvestimentos", []byte(launcher), 0o755},
		{appDir + "Contents/MacOS/invest-tracker-arm64", arm64, 0o755},
		{appDir + "Contents/MacOS/invest-tracker-amd64", amd64, 0o755},
		{appDir + "Contents/Resources/", nil, fs.ModeDir | 0o755},
		{appDir + "Contents/Resources/servidor.command", []byte(server), 0o755},
		{appDir + "Contents/Resources/AppIcon.icns", icon, 0o644},
		{"LEME.txt", []byte(readme), 0o644},
		{"README.txt", []byte(readmeEN), 0o644},
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		if err := add(zw, e); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := os.MkdirAll("dist-installers", 0o755); err != nil {
		return err
	}
	out := filepath.Join("dist-installers", "ControlInvestimentos-"+version+"-macos.zip")
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Println("Xerado", out)
	return nil
}

// add engade unha entrada co seu modo Unix: sen o bit de execución, macOS non
// podería abrir a app nin os seus binarios.
func add(zw *zip.Writer, e entry) error {
	h := &zip.FileHeader{Name: e.name, Method: zip.Deflate, Modified: modTime}
	if e.mode.IsDir() {
		h.Method = zip.Store
	}
	h.SetMode(e.mode)
	w, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	_, err = w.Write(e.data)
	return err
}

// readText le un ficheiro de texto con finais de liña LF, aínda que no
// repositorio estea con CRLF: con CRLF os scripts fallan en macOS.
func readText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n"), nil
}
