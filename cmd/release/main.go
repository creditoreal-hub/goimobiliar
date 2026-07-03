package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func run(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func pick(options []string) int {
	selected := 0

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	render := func() {
		fmt.Printf("\r\033[%dA", len(options))
		for i, opt := range options {
			if i == selected {
				fmt.Printf("\r  \033[32m▶  %s\033[0m\n", opt)
			} else {
				fmt.Printf("\r     %s\n", opt)
			}
		}
	}

	for range options {
		fmt.Println()
	}
	render()

	buf := make([]byte, 3)
	for {
		n, _ := os.Stdin.Read(buf)
		if n == 0 {
			continue
		}

		switch {
		case buf[0] == 13 || buf[0] == 10: // Enter
			fmt.Println()
			return selected
		case buf[0] == 3: // Ctrl+C
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Println("\n  Cancelado.")
			os.Exit(0)
		case n == 3 && buf[0] == 27 && buf[1] == 91: // ESC [ ...
			switch buf[2] {
			case 65: // Up
				if selected > 0 {
					selected--
				}
			case 66: // Down
				if selected < len(options)-1 {
					selected++
				}
			}
		}
		render()
	}
}

func main() {
	latest, err := run("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		latest = "v0.0.0"
	}

	version := strings.TrimPrefix(latest, "v")
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		fmt.Println("Não foi possível interpretar a versão da tag:", latest)
		os.Exit(1)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	fmt.Printf("\n  Versão atual: %s\n\n  Selecione o tipo de release:\n", latest)

	options := []string{
		fmt.Sprintf("major  ->  v%d.0.0", major+1),
		fmt.Sprintf("minor  ->  v%d.%d.0", major, minor+1),
		fmt.Sprintf("patch  ->  v%d.%d.%d", major, minor, patch+1),
	}

	choice := pick(options)

	var newVersion string
	switch choice {
	case 0:
		newVersion = fmt.Sprintf("v%d.0.0", major+1)
	case 1:
		newVersion = fmt.Sprintf("v%d.%d.0", major, minor+1)
	case 2:
		newVersion = fmt.Sprintf("v%d.%d.%d", major, minor, patch+1)
	}

	fmt.Printf("\n  Criar tag e enviar %s? [s/N]: ", newVersion)

	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	term.Restore(int(os.Stdin.Fd()), oldState)

	if buf[0] != 's' && buf[0] != 'S' {
		fmt.Println("\n  Cancelado.")
		os.Exit(0)
	}
	fmt.Println()

	if _, err := run("git", "tag", newVersion); err != nil {
		fmt.Println("  Falha ao criar a tag:", err)
		os.Exit(1)
	}

	if _, err := run("git", "push", "origin", newVersion); err != nil {
		fmt.Println("  Falha ao enviar a tag:", err)
		os.Exit(1)
	}

	fmt.Printf("\n  Release %s concluído!\n\n", newVersion)
}
