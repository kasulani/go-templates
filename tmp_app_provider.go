package app

import (
	"github.com/defval/di"
    {{range .imports}}
	"{{.}}"
	{{- end}}
)

func provideCliCommands() di.Option {
	return di.Options(
		di.Provide(startRootCommand),
	)
}

func provideSubCommands() di.Option {
	return di.Options(
	    {{- if .has.restAPI}}
		di.Provide(startAPIServerCommand, di.As(new(subCommand))),
		{{- end}}
	)
}

func provideTelemetry() di.Option {
    return di.Options(
        di.Provide(telemetry.NewInstrumentation),
        di.Provide(telemetry.NewMetricsRegistry),
    )
}

{{if .has.restAPI}}
func provideRESTAPIEndpoints() di.Option {
	return di.Options(
		di.Provide(rest.NewIndexEndpoint, di.As(new(rest.Endpoint))),
		di.Provide(rest.NewDocsEndpoint, di.As(new(rest.Endpoint))),
		di.Provide(rest.NewStatusEndpoint, di.As(new(rest.Endpoint))),
		di.Provide(rest.NewExampleEndpoint, di.As(new(rest.Endpoint))),
	)
}
{{end}}
{{- if .has.database}}
func provideDatabase() di.Option {
	return di.Options(
		di.Provide(database.NewDatabase),
		di.Provide(database.NewHealthChecker),
	)
}

func provideRepositories() di.Option {
    // todo: provide your repositories here
	return di.Options()
}
{{- end}}
{{if .has.httpClient}}
func provideHTTPClients() di.Option {
	return di.Options(
		di.Provide(httpclient.NewExampleClient),
	)
}
{{- end}}


