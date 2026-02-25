package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	stdio := flag.Bool("stdio", true, "Use stdio transport (default)")
	debug := flag.String("debug", "", "Start HTTP debug endpoint on addr (e.g. :9999)")
	version := flag.Bool("version", false, "Print version")
	formatFile := flag.String("format", "", "Format an ASON file and print to stdout")
	compressFile := flag.String("compress", "", "Compress an ASON file and print to stdout")
	flag.Parse()

	if *version {
		fmt.Println("ason-lsp v0.1.0")
		os.Exit(0)
	}

	if *formatFile != "" {
		data, err := os.ReadFile(*formatFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(Format(string(data)))
		return
	}

	if *compressFile != "" {
		var data []byte
		var err error
		if *compressFile == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(*compressFile)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(Compress(string(data)))
		return
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
