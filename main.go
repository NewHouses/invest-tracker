package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"invest-tracker/internal/addasset"
	"invest-tracker/internal/adddividend"
	"invest-tracker/internal/addresult"
	"invest-tracker/internal/addtransaction"
	"invest-tracker/internal/addtxbatch"
	"invest-tracker/internal/addtxmonth"
	"invest-tracker/internal/clearmonth"
	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/deleteasset"
	"invest-tracker/internal/deletedividend"
	"invest-tracker/internal/deleteresult"
	"invest-tracker/internal/deletetransaction"
	"invest-tracker/internal/editasset"
	"invest-tracker/internal/edittransaction"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
	"invest-tracker/internal/repartoaporte"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewassethistory"
	"invest-tracker/internal/viewcharts"
	"invest-tracker/internal/viewprojection"
	"invest-tracker/internal/viewreport"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtotalreport"
	"invest-tracker/internal/viewtransactions"
	"invest-tracker/internal/viewtypehistory"
	"invest-tracker/internal/viewtypereport"
	"invest-tracker/internal/web"
	"invest-tracker/internal/welcome"
)

// dbPath é a base de datos da CLI, relativa ao cartafol desde o que se executa.
// Os lanzadores (.cmd en Windows, .command en macOS) entran antes no cartafol
// do programa.
const dbPath = "./investimentos.db"

// runOp executa unha operación e amaña o erro: se é ErrCancelled imprime un
// aviso suave; calquera outro erro vai a stderr cun prefixo descritivo.
func runOp(label string, op func() error) {
	if err := op(); err != nil {
		if errors.Is(err, prompts.ErrCancelled) {
			i18n.Fprintln(os.Stdout, "↷ Operación cancelada. Volvendo ao menú.")
			return
		}
		// Os erros fixos (p.e. os do dominio) están en galego: tradúcense aquí.
		i18n.Fprintf(os.Stderr, "⚠ erro %s: %s\n", label, i18n.T(err.Error()))
	}
}

// initLanguage escolle o idioma da CLI e do servidor web: o que se gardou desde
// o menú ou, se non hai ningún, o do sistema.
func initLanguage() {
	lang := i18n.Detect()
	if saved, ok := i18n.LoadPreference(); ok {
		lang = saved
	}
	i18n.SetCurrent(lang)
}

// chooseLanguage pregunta o idioma e gárdao para as próximas veces.
func chooseLanguage(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("  [1] Galego")
	fmt.Println("  [2] English")
	fmt.Print("> ")
	line, eof := readMenuLine(reader)
	if eof {
		return
	}
	var lang i18n.Lang
	switch strings.TrimSpace(line) {
	case "1":
		lang = i18n.Galician
	case "2":
		lang = i18n.English
	default:
		fmt.Println("⚠", i18n.T(welcome.ErrInvalidOption.Error()))
		return
	}
	i18n.SetCurrent(lang)
	if err := i18n.SavePreference(lang); err != nil {
		i18n.Fprintf(os.Stderr, "⚠ non se puido gardar o idioma: %v\n", err)
	}
	i18n.Fprintln(os.Stdout, "✓ Idioma: galego")
}

func dispatch(catKey, opKey int, reader *bufio.Reader, s *store.Store) {
	switch catKey {
	case 1: // Operacións con activos
		switch opKey {
		case 1:
			runOp(i18n.T("engadindo activo"), func() error { return addasset.Run(reader, os.Stdout, s) })
		case 2:
			runOp(i18n.T("editando activo"), func() error { return editasset.Run(reader, os.Stdout, s) })
		case 3:
			runOp(i18n.T("eliminando activo"), func() error { return deleteasset.Run(reader, os.Stdout, s) })
		}
	case 2: // Operacións con transaccións
		switch opKey {
		case 1:
			runOp(i18n.T("engadindo transacción"), func() error { return addtransaction.Run(reader, os.Stdout, s) })
		case 2:
			runOp(i18n.T("engadindo varias transaccións"), func() error { return addtxbatch.Run(reader, os.Stdout, s) })
		case 3:
			runOp(i18n.T("engadindo transaccións do mes"), func() error { return addtxmonth.Run(reader, os.Stdout, s) })
		case 4:
			runOp(i18n.T("repartindo aporte mensual"), func() error { return repartoaporte.Run(reader, os.Stdout, s) })
		case 5:
			runOp(i18n.T("editando transacción"), func() error { return edittransaction.Run(reader, os.Stdout, s) })
		case 6:
			runOp(i18n.T("eliminando transacción"), func() error { return deletetransaction.Run(reader, os.Stdout, s) })
		}
	case 3: // Operacións de resultados
		switch opKey {
		case 1:
			runOp(i18n.T("engadindo resultado"), func() error { return addresult.Run(reader, os.Stdout, s) })
		case 2:
			runOp(i18n.T("engadindo dividendo"), func() error { return adddividend.Run(reader, os.Stdout, s) })
		case 3:
			runOp(i18n.T("pechando o mes"), func() error { return closemonth.Run(reader, os.Stdout, s) })
		case 4:
			runOp(i18n.T("eliminando resultado"), func() error { return deleteresult.Run(reader, os.Stdout, s) })
		case 5:
			runOp(i18n.T("eliminando dividendo"), func() error { return deletedividend.Run(reader, os.Stdout, s) })
		case 6:
			runOp(i18n.T("limpando o mes"), func() error { return clearmonth.Run(reader, os.Stdout, s) })
		}
	case 4: // Informes
		switch opKey {
		case 1:
			runOp(i18n.T("xerando historial"), func() error { return viewassethistory.Run(reader, os.Stdout, s) })
		case 2:
			runOp(i18n.T("xerando historial por tipo"), func() error { return viewtypehistory.Run(reader, os.Stdout, s) })
		case 3:
			runOp(i18n.T("listando transaccións"), func() error { return viewtransactions.Run(reader, os.Stdout, s) })
		case 4:
			runOp(i18n.T("xerando informe"), func() error { return viewreport.Run(reader, os.Stdout, s) })
		case 5:
			runOp(i18n.T("xerando informe por tipo"), func() error { return viewtypereport.Run(reader, os.Stdout, s) })
		case 6:
			runOp(i18n.T("xerando informe total"), func() error { return viewtotalreport.Run(reader, os.Stdout, s) })
		case 7:
			runOp(i18n.T("xerando reporte histórico completo"), func() error { return viewtotalhistory.Run(reader, os.Stdout, s) })
		case 8:
			runOp(i18n.T("debuxando gráficas"), func() error { return viewcharts.Run(reader, os.Stdout, s) })
		case 9:
			runOp(i18n.T("calculando a proxección"), func() error { return viewprojection.Run(reader, os.Stdout, s) })
		}
	}
}

// readMenuLine reads a line from stdin. Returns (line, eof). On EOF prints a
// trailing newline and signals the caller to exit cleanly.
func readMenuLine(reader *bufio.Reader) (string, bool) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Println()
			return "", true
		}
		i18n.Fprintf(os.Stderr, "erro lendo entrada: %v\n", err)
		return "", true
	}
	return line, false
}

func main() {
	initLanguage()
	if len(os.Args) > 1 && os.Args[1] == "web" {
		os.Exit(web.Main(os.Args[2:]))
	}

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(i18n.Sprintf("non se pode abrir a base de datos: %v", err))
	}
	defer s.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println(welcome.WelcomeMessage())
	if abs, err := filepath.Abs(dbPath); err == nil {
		fmt.Println(i18n.Sprintf("(Datos: %s)", abs))
	}

	for {
		topOpts := welcome.TopOptions()
		fmt.Println()
		fmt.Print(welcome.Render(topOpts))
		fmt.Print("> ")
		line, eof := readMenuLine(reader)
		if eof {
			return
		}
		topSel, err := welcome.Select(topOpts, line)
		if err != nil {
			fmt.Println("⚠", i18n.T(err.Error()))
			continue
		}
		if topSel.Key == 0 {
			i18n.Fprintln(os.Stdout, "Ata logo!")
			return
		}
		if topSel.Key == welcome.LanguageKey {
			chooseLanguage(reader)
			continue
		}

		// Bucle do submenú: o usuario pode encadear varias operacións dentro
		// dunha mesma categoría ata premer 0 (Voltar).
		for {
			subOpts := welcome.CategoryOptions(topSel.Key)
			fmt.Println()
			fmt.Printf("--- %s ---\n", topSel.Label)
			fmt.Print(welcome.Render(subOpts))
			fmt.Print("> ")
			line, eof := readMenuLine(reader)
			if eof {
				return
			}
			subSel, err := welcome.Select(subOpts, line)
			if err != nil {
				fmt.Println("⚠", i18n.T(err.Error()))
				continue
			}
			if subSel.Key == 0 {
				break
			}
			dispatch(topSel.Key, subSel.Key, reader, s)
		}
	}
}
