package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"invest-tracker/internal/welcome"
)

func main() {
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
      fmt.Println("⚠️", err)
      continue
    }
    if opt.Key == 0 {
      fmt.Println("Ata logo!")
      return
    }
    switch opt.Key {
    default:
      fmt.Printf("Seleccionaches: %s (placeholder, aínda non implementado)\n", opt.Label)
    }
  }
}