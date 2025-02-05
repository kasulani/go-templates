package app

import (
	"context"
	"log"
	"os"

	"github.com/defval/di"
	"github.com/kelseyhightower/envconfig"
	{{range .imports}}
	"{{.}}"
	{{- end}}
)

type config struct {
	RestAPIServerAddress string `envconfig:"REST_API_SERVER_ADDRESS" default:"0.0.0.0:8000"` //nolint:lll
}

func newConfig() *config {
	cfg := new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("failed to load configuration: %q", err)
	}

	return cfg
}

func registerSubCommands(root *rootCommand, subCommands subCommands) {
	for _, subCommand := range subCommands {
		subCommand.AddTo(root)
	}
}

// Run is an app entry method.
func Run(root *rootCommand) error {
	return root.Execute()
}
{{if .has.database}}
// TerminateConnections will close connections to app dependencies.
func TerminateConnections(db *database.Connection) {
	err := db.Close()
	if err != nil {
		log.Fatalf("failed to close database connection: %q", err)
	}
}
{{end}}
// Container is a dependency injection container.
func Container() *di.Container {
	if os.Getenv("LOG_LEVEL") == "debug" {
		di.SetTracer(&di.StdTracer{})
	}

	c, err := di.New(
		di.Provide(context.Background),
		provideTelemetry(),
        provideCliCommands(),
        provideSubCommands(),
        di.Invoke(registerSubCommands),
		{{- if .has.restAPI}}
		provideRESTAPIEndpoints(),
		di.Provide(rest.NewAPIServer),
		di.Invoke(rest.RegisterAPIEndpoints),
		{{- end}}
		{{- if .has.httpClient}}
		provideHTTPClients(),
		{{- end}}
        {{- if .has.database}}
		provideDatabase(),
		provideRepositories(),
		{{- end}}
	)
	if err != nil {
		log.Fatalf("failed to create DI container: %q", err)
	}

	return c
}
