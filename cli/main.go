package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/clawhost/core/app"
)

func main() {
	if len(os.Args) > 1 {
		if err := runCLI(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			printCLIUsage()
			os.Exit(1)
		}
		return
	}

	var (
		port = flag.String("port", "8080", "Port to run the server on")
	)
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}

	server, err := app.InitializeCoreServer(*port)
	if err != nil {
		panic(err)
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}
