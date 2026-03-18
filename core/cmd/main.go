package main

import (
	"flag"
	"os"

	"github.com/yourusername/clawhost/core/app"
)

func main() {
	var (
		port = flag.String("port", "8080", "Port to run the server on")
	)
	flag.Parse()

	// Override port from environment if set
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
