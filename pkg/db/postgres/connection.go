package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver for database/sql
)

// PostgresURLEnvVar represents env variable for postgres connection url
const PostgresURLEnvVar = "POSTGRES_URL"

// NewConnection starts a new postgres db connection
func NewConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return db, nil
}
