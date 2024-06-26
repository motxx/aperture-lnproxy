package aperturedb

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	postgres_migrate "github.com/golang-migrate/migrate/v4/database/postgres"
	// Import the file source to register it with the migrate library.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/motxx/aperture-lnproxy/aperture/aperturedb/sqlc"
	"github.com/stretchr/testify/require"
)

const (
	dsnTemplate = "postgres://%v:%v@%v:%d/%v?sslmode=%v"
)

var (
	// DefaultPostgresFixtureLifetime is the default maximum time a Postgres
	// test fixture is being kept alive. After that time the docker
	// container will be terminated forcefully, even if the tests aren't
	// fully executed yet. So this time needs to be chosen correctly to be
	// longer than the longest expected individual test run time.
	DefaultPostgresFixtureLifetime = 10 * time.Minute

	// defaultMaxConns is the number of permitted active and idle
	// connections. We want to limit this so it isn't unlimited. We use the
	// same value for the number of idle connections as, this can speed up
	// queries given a new connection doesn't need to be established each
	// time.
	defaultMaxConns = 25

	// connIdleLifetime is the amount of time a connection can be idle.
	connIdleLifetime = 5 * time.Minute
)

// PostgresConfig holds the postgres database configuration.
type PostgresConfig struct {
	SkipMigrations     bool   `long:"skipmigrations" description:"Skip applying migrations on startup."`
	Host               string `long:"host" description:"Database server hostname."`
	Port               int    `long:"port" description:"Database server port."`
	User               string `long:"user" description:"Database user."`
	Password           string `long:"password" description:"Database user's password."`
	DBName             string `long:"dbname" description:"Database name to use."`
	MaxOpenConnections int32  `long:"maxconnections" description:"Max open connections to keep alive to the database server."`
	RequireSSL         bool   `long:"requiressl" description:"Whether to require using SSL (mode: require) when connecting to the server."`
}

// DSN returns the dns to connect to the database.
func (s *PostgresConfig) DSN(hidePassword bool) string {
	var sslMode = "disable"
	if s.RequireSSL {
		sslMode = "require"
	}

	password := s.Password
	if hidePassword {
		// Placeholder used for logging the DSN safely.
		password = "****"
	}

	return fmt.Sprintf(dsnTemplate, s.User, password, s.Host, s.Port,
		s.DBName, sslMode)
}

// PostgresStore is a database store implementation that uses a Postgres
// backend.
type PostgresStore struct {
	cfg *PostgresConfig

	*BaseDB
}

// NewPostgresStore creates a new store that is backed by a Postgres database
// backend.
func NewPostgresStore(cfg *PostgresConfig) (*PostgresStore, error) {
	log.Infof("Using SQL database '%s'", cfg.DSN(true))

	rawDB, err := sql.Open("pgx", cfg.DSN(false))
	if err != nil {
		return nil, err
	}

	maxConns := defaultMaxConns
	if cfg.MaxOpenConnections > 0 {
		maxConns = int(cfg.MaxOpenConnections)
	}

	rawDB.SetMaxOpenConns(maxConns)
	rawDB.SetMaxIdleConns(maxConns)
	rawDB.SetConnMaxLifetime(connIdleLifetime)

	if !cfg.SkipMigrations {
		// Now that the database is open, populate the database with
		// our set of schemas based on our embedded in-memory file
		// system.
		driver, err := postgres_migrate.WithInstance(
			rawDB, &postgres_migrate.Config{},
		)
		if err != nil {
			return nil, err
		}

		postgresFS := newReplacerFS(sqlSchemas, map[string]string{
			"BLOB":                "BYTEA",
			"INTEGER PRIMARY KEY": "SERIAL PRIMARY KEY",
			"TIMESTAMP":           "TIMESTAMP WITHOUT TIME ZONE",
		})

		err = applyMigrations(
			postgresFS, driver, "sqlc/migrations", cfg.DBName,
		)
		if err != nil {
			return nil, err
		}
	}

	queries := sqlc.New(rawDB)

	return &PostgresStore{
		cfg: cfg,
		BaseDB: &BaseDB{
			DB:      rawDB,
			Queries: queries,
		},
	}, nil
}

// NewTestPostgresDB is a helper function that creates a Postgres database for
// testing.
func NewTestPostgresDB(t *testing.T) *PostgresStore {
	t.Helper()

	t.Logf("Creating new Postgres DB for testing")

	sqlFixture := NewTestPgFixture(t, DefaultPostgresFixtureLifetime)
	store, err := NewPostgresStore(sqlFixture.GetConfig())
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlFixture.TearDown(t)
	})

	return store
}
