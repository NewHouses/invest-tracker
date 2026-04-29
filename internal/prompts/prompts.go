package prompts

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ReadLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func Amount(r *bufio.Reader, w io.Writer) (float64, error) {
	for {
		fmt.Fprint(w, "Cantidade (USD): ")
		line, err := ReadLine(r)
		if err != nil {
			return 0, err
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil && v > 0 {
			return v, nil
		}
		fmt.Fprintln(w, "⚠ Cantidade non válida, debe ser un número maior ca 0")
	}
}

func Month(r *bufio.Reader, w io.Writer) (int, error) {
	for {
		fmt.Fprint(w, "Mes (1-12): ")
		line, err := ReadLine(r)
		if err != nil {
			return 0, err
		}
		v, perr := strconv.Atoi(line)
		if perr == nil && v >= 1 && v <= 12 {
			return v, nil
		}
		fmt.Fprintln(w, "⚠ Mes non válido, debe estar entre 1 e 12")
	}
}

func Year(r *bufio.Reader, w io.Writer) (int, error) {
	for {
		fmt.Fprint(w, "Ano: ")
		line, err := ReadLine(r)
		if err != nil {
			return 0, err
		}
		v, perr := strconv.Atoi(line)
		if perr == nil && v >= 1900 && v <= 2100 {
			return v, nil
		}
		fmt.Fprintln(w, "⚠ Ano non válido")
	}
}
