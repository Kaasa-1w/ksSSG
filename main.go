package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"ks-ssg/internal/site"
)

// main runs the site generator with command-line options and default settings.
func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	isBuildCommand := false
	if len(args) > 0 {
		if args[0] == "build" {
			isBuildCommand = true
			args = args[1:]
		} else if args[0][0] != '-' {
			return fmt.Errorf("unknown command %q (available command: build)", args[0])
		}
	}

	config := site.DefaultConfig()
	flags := flag.NewFlagSet("ks-ssg", flag.ContinueOnError)
	flags.StringVar(&config.SourceDir, "source-dir", config.SourceDir, "directory containing Markdown source files")
	flags.StringVar(&config.OutputDir, "output-dir", config.OutputDir, "directory for generated HTML files")
	flags.StringVar(&config.LayoutDir, "layout-dir", config.LayoutDir, "directory containing HTML layouts")
	flags.Usage = func() {
		if isBuildCommand {
			fmt.Fprintln(flags.Output(), "Usage: ks-ssg build [options]")
		} else {
			fmt.Fprintln(flags.Output(), "Usage: ks-ssg <command> [options]")
			fmt.Fprintln(flags.Output(), "Commands:")
			fmt.Fprintln(flags.Output(), "  build  Generate the site")
			fmt.Fprintln(flags.Output(), "Options also work without the build command.")
		}
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("unexpected positional arguments")
	}
	return site.Build(config)
}
