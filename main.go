package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

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
	"invest-tracker/internal/prompts"
	"invest-tracker/internal/repartoaporte"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewassethistory"
	"invest-tracker/internal/viewreport"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtotalreport"
	"invest-tracker/internal/viewtransactions"
	"invest-tracker/internal/viewtypehistory"
	"invest-tracker/internal/viewtypereport"
	"invest-tracker/internal/welcome"
)

// runOp executa unha operación e amaña o erro: se é ErrCancelled imprime un
// aviso suave; calquera outro erro vai a stderr cun prefixo descritivo.
func runOp(label string, op func() error) {
	if err := op(); err != nil {
		if errors.Is(err, prompts.ErrCancelled) {
			fmt.Println("↷ Operación cancelada. Volvendo ao menú.")
			return
		}
		fmt.Fprintln(os.Stderr, "⚠ erro "+label+":", err)
	}
}

func dispatch(catKey, opKey int, reader *bufio.Reader, s *store.Store) {
	switch catKey {
	case 1: // Operacións con activos
		switch opKey {
		case 1:
			runOp("engadindo activo", func() error { return addasset.Run(reader, os.Stdout, s) })
		case 2:
			runOp("editando activo", func() error { return editasset.Run(reader, os.Stdout, s) })
		case 3:
			runOp("eliminando activo", func() error { return deleteasset.Run(reader, os.Stdout, s) })
		}
	case 2: // Operacións con transaccións
		switch opKey {
		case 1:
			runOp("engadindo transacción", func() error { return addtransaction.Run(reader, os.Stdout, s) })
		case 2:
			runOp("engadindo varias transaccións", func() error { return addtxbatch.Run(reader, os.Stdout, s) })
		case 3:
			runOp("engadindo transaccións do mes", func() error { return addtxmonth.Run(reader, os.Stdout, s) })
		case 4:
			runOp("repartindo aporte mensual", func() error { return repartoaporte.Run(reader, os.Stdout, s) })
		case 5:
			runOp("editando transacción", func() error { return edittransaction.Run(reader, os.Stdout, s) })
		case 6:
			runOp("eliminando transacción", func() error { return deletetransaction.Run(reader, os.Stdout, s) })
		}
	case 3: // Operacións de resultados
		switch opKey {
		case 1:
			runOp("engadindo resultado", func() error { return addresult.Run(reader, os.Stdout, s) })
		case 2:
			runOp("engadindo dividendo", func() error { return adddividend.Run(reader, os.Stdout, s) })
		case 3:
			runOp("pechando o mes", func() error { return closemonth.Run(reader, os.Stdout, s) })
		case 4:
			runOp("eliminando resultado", func() error { return deleteresult.Run(reader, os.Stdout, s) })
		case 5:
			runOp("eliminando dividendo", func() error { return deletedividend.Run(reader, os.Stdout, s) })
		case 6:
			runOp("limpando o mes", func() error { return clearmonth.Run(reader, os.Stdout, s) })
		}
	case 4: // Informes
		switch opKey {
		case 1:
			runOp("xerando historial", func() error { return viewassethistory.Run(reader, os.Stdout, s) })
		case 2:
			runOp("xerando historial por tipo", func() error { return viewtypehistory.Run(reader, os.Stdout, s) })
		case 3:
			runOp("listando transaccións", func() error { return viewtransactions.Run(reader, os.Stdout, s) })
		case 4:
			runOp("xerando informe", func() error { return viewreport.Run(reader, os.Stdout, s) })
		case 5:
			runOp("xerando informe por tipo", func() error { return viewtypereport.Run(reader, os.Stdout, s) })
		case 6:
			runOp("xerando informe total", func() error { return viewtotalreport.Run(reader, os.Stdout, s) })
		case 7:
			runOp("xerando reporte histórico completo", func() error { return viewtotalhistory.Run(reader, os.Stdout, s) })
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
		fmt.Fprintln(os.Stderr, "erro lendo entrada:", err)
		return "", true
	}
	return line, false
}

func main() {
	s, err := store.Open("./investimentos.db")
	if err != nil {
		log.Fatalf("non se pode abrir a base de datos: %v", err)
	}
	defer s.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println(welcome.WelcomeMessage())

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
			fmt.Println("⚠", err)
			continue
		}
		if topSel.Key == 0 {
			fmt.Println("Ata logo!")
			return
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
				fmt.Println("⚠", err)
				continue
			}
			if subSel.Key == 0 {
				break
			}
			dispatch(topSel.Key, subSel.Key, reader, s)
		}
	}
}
