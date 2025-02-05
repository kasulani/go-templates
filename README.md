# go-templates

A collection of Go project templates for use with [Fundi](https://github.com/kasulani/go-fundi). These templates help standardize project structure and streamline the generation of Go applications, including REST APIs, database integrations, and telemetry setups.

## Features
✅ Predefined templates for Go projects  
✅ Supports Fundi-powered project scaffolding  
✅ Customizable placeholders for dynamic configurations  
✅ Includes REST, database, and telemetry components

## 📌 Prerequisites

Before using these templates, ensure you have:

- Go installed ([Download Go](https://go.dev/dl/))
- Fundi installed ([GitHub - kasulani/go-fundi](https://github.com/kasulani/go-fundi))
- A valid `.fundi.yml` file to guide the generation process

---

## 🚀 How to Use the Templates

### Create a `.fundi.yml` configuration file

Edit the `config.yml` file to match your project structure; change the `output`, `templates`, and `values` fields to
match your project's requirements. The `variables` field contains variables that can be used in the `values` file.
Update the `project` and `repository` values accordingly. The `directories` field contains the project structure. Give
an appropriate name to the root directory. Each directory can contain files and subdirectories. The `files` field
contains the files to be created in the directory. The `template` field specifies the template file to use. The
`directories` field contains subdirectories of the directory. The `name` field specifies the name of the directory or
file.

```yaml
metadata:
  output: "." # location where the project files will be created
  templates: "./go-templates" # location of the template files
  values: "./go-templates/values.yml" # values file
  variables: # variables to be used in the values file
    project: simpleAPIServer
    repository: github.com/kasulani
directories:
  - name: simpleAPIServer # root directory of your project
    files:
      - name: README.md
      - name: Makefile
        template: makefile
      - name: docker-compose.yml
        template: docker_compose
    directories:
      - name: cmd
        directories:
          - name: app
            files:
              - name: main.go
                template: tmp_main.go
      - name: dockerfiles
        directories:
          - name: development
            files:
              - name: Dockerfile
                template: dockerfile
      - name: internal # internal package
        directories:
          - name: app # app package
            files:
              - name: app.go
                template: tmp_app.go
              - name: command.go
                template: tmp_app_command.go
              - name: provider.go
                template: tmp_app_provider.go
          - name: database # database package
            files:
              - name: database.go
                template: tmp_database.go
          - name: logging # logging package
            files:
              - name: logger.go
                template: tmp_logger.go
          - name: repository # repository package
            files:
              - name: repository.go
                template: tmp_repository.go
              - name: factory.go
                template: tmp_repository_factory.go
          - name: httpclient # httpclient package
            files:
              - name: factory.go
                template: tmp_httpclient_factory.go
              - name: example.go
                template: tmp_httpclient_example.go
              - name: middleware.go
                template: tmp_httpclient_middleware.go
              - name: interface.go
                template: tmp_httpclient_interface.go
              - name: type.go
                template: tmp_httpclient_type.go
          - name: rest # rest package
            files:
              - name: factory.go
                template: tmp_rest_factory.go
              - name: interface.go
                template: tmp_rest_interface.go
              - name: middleware.go
                template: tmp_rest_middleware.go
              - name: endpoints.go
                template: tmp_rest_endpoints.go
              - name: server.go
                template: tmp_rest_server.go
          - name: telemetry # telemetry package
            files:
              - name: factory.go
                template: tmp_telemetry_factory.go
              - name: interface.go
                template: tmp_telemetry_interface.go
              - name: telemetry.go
                template: tmp_telemetry.go
```

### Clone the Templates Repository

```bash
git clone https://github.com/your-org/templates-repo.git
cd templates-repo
```

### Generate a New Project

```bash
fundi generate -f .fundi.yml
```

Change to the project directory and initialize a new git repository and go module.

```bash
go mod init
go mod tidy
go mod vendor
git init
```

### Inspect and edit the generated files

Inspect the generated files and make any necessary changes to match your project requirements. Remove any unnecessary
files and directories.

#### Dockerfile

Edit the `Dockerfile` to match your project requirements. Make sure the image tag matches the project requirements. Make
sure the working directory matches the one in the docker-compose file.

#### docker-compose.yml

Edit the `dev` container `command` to match your project requirements. For example,
`CompileDaemon -build='make install-all' -command='app start-api'`
if the port in the generated file is already in use, change it to a free port.
