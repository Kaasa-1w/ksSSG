package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// main runs the site generator CLI.
func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	return runAt(args, ".")
}

func runAt(args []string, directory string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printUsage()
		return nil
	}

	switch args[0] {
	case "init":
		if len(args) != 1 {
			return errors.New("init does not accept arguments")
		}
		return initializeProject(directory)
	case "build":
		if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
			fmt.Println("Usage: ks-ssg build")
			return nil
		}
		if len(args) != 1 {
			return errors.New("build does not accept arguments; configure directories in ks-ssg.json")
		}
		return buildProject(directory)
	default:
		return fmt.Errorf("unknown command %q (available commands: init, build)", args[0])
	}
}

func printUsage() {
	fmt.Println("Usage: ks-ssg <command>")
	fmt.Println("Commands:")
	fmt.Println("  init   Create a ks-ssg.json configuration in the current directory")
	fmt.Println("  build  Generate the site using ks-ssg.json")
}
