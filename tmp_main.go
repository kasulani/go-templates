package main

import (
	"log"
	"os"
    {{range .imports}}
	"{{.}}"
	{{- end}}
)

func main() {
	container := app.Container()

	if err := container.Invoke(app.Run); err != nil {
		log.Printf("failed to start application: %q\n", err)
		os.Exit(1)
	}

	if err := container.Invoke(app.TerminateConnections); err != nil {
		log.Printf("failed to terminate connections: %q\n", err)
		os.Exit(1)
	}

	defer container.Cleanup()
}