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
	"invest-tracker/internal/clearmonth"
	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/deleteasset"
	"invest-tracker/internal/deletedividend"
	"invest-tracker/internal/deleteresult"
	"invest-tracker/internal/deletetransaction"
	"invest-tracker/internal/editasset"
	"invest-tracker/internal/prompts"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewassetgeneral"
	"invest-tracker/internal/viewassethistory"
	"invest-tracker/internal/viewreport"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtotalreport"
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

func main() {
	s, err := store.Open("./investimentos.db")
	if err != nil {
		log.Fatalf("non se pode abrir a base de datos: %v", err)
	}
	defer s.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println(welcome.WelcomeMessage())
	for {
		fmt.Println()
		fmt.Print(welcome.Render(welcome.Options()))
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return
			}
			fmt.Fprintln(os.Stderr, "erro lendo entrada:", err)
			return
		}
		opt, err := welcome.Select(welcome.Options(), line)
		if err != nil {
			fmt.Println("⚠", err)
			continue
		}
		if opt.Key == 0 {
			fmt.Println("Ata logo!")
			return
		}
		switch opt.Key {
		case 1:
			runOp("engadindo activo", func() error { return addasset.Run(reader, os.Stdout, s) })
		case 2:
			runOp("editando activo", func() error { return editasset.Run(reader, os.Stdout, s) })
		case 3:
			runOp("eliminando activo", func() error { return deleteasset.Run(reader, os.Stdout, s) })
		case 4:
			runOp("engadindo transacción", func() error { return addtransaction.Run(reader, os.Stdout, s) })
		case 5:
			runOp("engadindo varias transaccións", func() error { return addtxbatch.Run(reader, os.Stdout, s) })
		case 6:
			runOp("eliminando transacción", func() error { return deletetransaction.Run(reader, os.Stdout, s) })
		case 7:
			runOp("pechando o mes", func() error { return closemonth.Run(reader, os.Stdout, s) })
		case 8:
			runOp("limpando o mes", func() error { return clearmonth.Run(reader, os.Stdout, s) })
		case 9:
			runOp("engadindo resultado", func() error { return addresult.Run(reader, os.Stdout, s) })
		case 10:
			runOp("eliminando resultado", func() error { return deleteresult.Run(reader, os.Stdout, s) })
		case 11:
			runOp("engadindo dividendo", func() error { return adddividend.Run(reader, os.Stdout, s) })
		case 12:
			runOp("eliminando dividendo", func() error { return deletedividend.Run(reader, os.Stdout, s) })
		case 13:
			runOp("xerando informe", func() error { return viewreport.Run(reader, os.Stdout, s) })
		case 14:
			runOp("xerando resultado xeral do activo", func() error { return viewassetgeneral.Run(reader, os.Stdout, s) })
		case 15:
			runOp("xerando informe por tipo", func() error { return viewtypereport.Run(reader, os.Stdout, s) })
		case 16:
			runOp("xerando resultado xeral", func() error { return viewtotalhistory.Run(reader, os.Stdout, s) })
		case 17:
			runOp("xerando informe total", func() error { return viewtotalreport.Run(reader, os.Stdout, s) })
		case 18:
			runOp("xerando historial", func() error { return viewassethistory.Run(reader, os.Stdout, s) })
		default:
			fmt.Printf("Seleccionaches: %s (placeholder, aínda non implementado)\n", opt.Label)
		}
	}
}
