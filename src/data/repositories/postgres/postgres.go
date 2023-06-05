package postgres_repository

import (
	"database/sql"
	"fmt"

	"github.com/oceano-dev/microservices-go-common/config"
)

func NewPostgresDatabase(config *config.Config) (*sql.DB, error) {
	conn := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s",
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Database,
		config.Postgres.SSLMode)

	return sql.Open("postgres", conn)
}
