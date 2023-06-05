package migrate

import (
	"fmt"
	"log"

	"github.com/oceano-dev/microservices-go-common/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(config *config.Config) {
	postgres := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.Database,
		config.Postgres.SSLMode)

	m, err := migrate.New(
		"file://sql",
		postgres)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Println(err)
	}

	fmt.Println("Migrations done!!!")
}
