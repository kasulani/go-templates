package app

import (
	"context"

	"github.com/spf13/cobra"
	{{range .imports}}
	"{{.}}"
	{{- end}}
)
type (
	// Command is cli command type.
	Command struct {
		*cobra.Command
	}

	// subCommand interface defines AddTo method.
	subCommand interface {
		AddTo(root *rootCommand)
	}

	// subCommands is a list of subCommand.
	subCommands []subCommand

	rootCommand    Command
	startAPIServer Command
)

func startRootCommand() *rootCommand {
	return &rootCommand{
		Command: &cobra.Command{
			Use:     "{{.rootCommand}}",
			Short:   "Use this command to manipulate the application",
			Long:    `Use this command to manipulate the application`,
			Version: "1.0.0",
		},
	}
}
{{if .has.restAPI}}
func startAPIServerCommand(
	ctx context.Context,
	apiServer *rest.APIServer,
) *startAPIServer {
	cfg := newConfig()
	return &startAPIServer{
		&cobra.Command{
			Use:     "start-api-server",
			Aliases: []string{"start-api"},
			Short:   "start REST API server",
			Long:    "This command starts REST API server",
			Run: func(cmd *cobra.Command, args []string) {
				rest.BootstrapAPIServer(ctx, apiServer, cfg.RestAPIServerAddress)
			},
		},
	}
}

func (startAPIServer *startAPIServer) AddTo(root *rootCommand) {
	root.AddCommand(startAPIServer.Command)
}
{{end}}
