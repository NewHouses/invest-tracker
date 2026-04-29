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
	"invest-tracker/internal/deletetransaction"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewassethistory"
	"invest-tracker/internal/viewreport"
	"invest-tracker/internal/welcome"
)

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
			if err := addasset.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro engadindo activo:", err)
			}
		case 2:
			if err := addtransaction.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro engadindo transacción:", err)
			}
		case 3:
			if err := deletetransaction.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro eliminando transacción:", err)
			}
		case 4:
			if err := addresult.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro engadindo resultado:", err)
			}
		case 5:
			if err := adddividend.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro engadindo dividendo:", err)
			}
		case 6:
			if err := viewreport.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro xerando informe:", err)
			}
		case 7:
			if err := viewassethistory.Run(reader, os.Stdout, s); err != nil {
				fmt.Fprintln(os.Stderr, "⚠ erro xerando historial:", err)
			}
		default:
			fmt.Printf("Seleccionaches: %s (placeholder, aínda non implementado)\n", opt.Label)
		}
	}
}
