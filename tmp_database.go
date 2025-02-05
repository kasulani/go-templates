package database

import (
	"context"
	"database/sql"
	"log"
	"time"
    {{if .uses.postgres}}
	_ "github.com/lib/pq" // Blank import to load and register the PostgreSQL driver.
	{{- end}}

	"github.com/alexliesenfeld/health"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"go.nhat.io/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	{{range .imports}}
    "{{.}}"
    {{- end}}
)

type (
	Connection struct {
		*sqlx.DB
	}

	config struct {
		DSN                string        `envconfig:"DATABASE_DSN" required:"true"`
		MaxOpenConns       int           `envconfig:"DATABASE_MAX_OPEN_CONNS" default:"50"`
		MaxIdleConns       int           `envconfig:"DATABASE_MAX_IDLE_CONNS" default:"50"`
		ConnMaxLifetime    time.Duration `envconfig:"DATABASE_CONN_MAX_LIFETIME" default:"30m"`
		ConnMaxIdleTimeout time.Duration `envconfig:"DATABASE_CONN_MAX_IDLE_TIMEOUT" default:"10m"`
	}
)

const (
    {{- if .uses.postgres}}
	dbDriver = "postgres"
	{{- end}}

	timeout = time.Second * 10
)

// NewDatabase returns a database Connection.
func NewDatabase(instrumentation *telemetry.Instrumentation) *Connection {
	dbCfg := newConfig()

	driver, err := otelsql.Register(
		dbDriver,
		otelsql.AllowRoot(),
		otelsql.TraceQueryWithoutArgs(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsAffected(),
		otelsql.WithDatabaseName("{{.databaseName}}"),
		otelsql.WithDefaultAttributes(semconv.ServiceName(instrumentation.ServiceName())),
	)
	if err != nil {
		log.Fatalf("failed to register an otelsql driver: %q", err)
	}

	db, err := sql.Open(driver, dbCfg.DSN)
	if err != nil {
		log.Fatalf("failed to open new database connection: %q", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %q", err)
	}

	db.SetMaxIdleConns(dbCfg.MaxIdleConns)
	db.SetMaxOpenConns(dbCfg.MaxOpenConns)
	db.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(dbCfg.ConnMaxIdleTimeout)

	if err := otelsql.RecordStats(db); err != nil {
		log.Fatalf("failed to record database statistics: %q", err)
	}

	return &Connection{sqlx.NewDb(db, dbDriver)}
}

// NewHealthChecker returns an instance of a database health checker.
func NewHealthChecker(database *Connection) health.Checker {
	return health.NewChecker(
		health.WithCheck(
			health.Check{
				Name: "database",
				Check: func(ctx context.Context) error {
					return database.PingContext(ctx)
				},
				Timeout: timeout,
			},
		),
	)
}

func newConfig() *config {
	cfg := new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("failed to load configuration: %q", err)
	}

	return cfg
}
