package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	stdio := flag.Bool("stdio", true, "Use stdio transport (default)")
	debug := flag.String("debug", "", "Start HTTP debug endpoint on addr (e.g. :9999)")
	version := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *version {
		fmt.Println("ason-lsp v0.1.0")
		os.Exit(0)
	}

	if *debug != "" {
		startHTTPDebug(*debug)
		fmt.Fprintf(os.Stderr, "[ason-lsp] HTTP debug at %s\n", *debug)
	}

	if *stdio {
		srv := NewServer(os.Stdin, os.Stdout)
		if err := srv.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "ason-lsp error: %v\n", err)
			os.Exit(1)
		}
	}
}
