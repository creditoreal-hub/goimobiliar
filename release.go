package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	latest, err := run("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		latest = "v0.0.0"
	}

	version := strings.TrimPrefix(latest, "v")
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		fmt.Println("Could not parse version from tag:", latest)
		os.Exit(1)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	fmt.Printf("\n  Current version: %s\n\n", latest)
	fmt.Printf("  Select release type:\n")
	fmt.Printf("    1) major  ->  v%d.0.0\n", major+1)
	fmt.Printf("    2) minor  ->  v%d.%d.0\n", major, minor+1)
	fmt.Printf("    3) patch  ->  v%d.%d.%d\n\n", major, minor, patch+1)
	fmt.Print("  Choice [1/2/3]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var newVersion string
	switch choice {
	case "1":
		newVersion = fmt.Sprintf("v%d.0.0", major+1)
	case "2":
		newVersion = fmt.Sprintf("v%d.%d.0", major, minor+1)
	case "3":
		newVersion = fmt.Sprintf("v%d.%d.%d", major, minor, patch+1)
	default:
		fmt.Println("  Invalid choice. Aborting.")
		os.Exit(1)
	}

	fmt.Printf("\n  Tag and push %s? [y/N]: ", newVersion)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("  Aborted.")
		os.Exit(0)
	}

	if _, err := run("git", "tag", newVersion); err != nil {
		fmt.Println("  Failed to create tag:", err)
		os.Exit(1)
	}

	if _, err := run("git", "push", "origin", newVersion); err != nil {
		fmt.Println("  Failed to push tag:", err)
		os.Exit(1)
	}

	fmt.Printf("\n  Released %s\n\n", newVersion)
}