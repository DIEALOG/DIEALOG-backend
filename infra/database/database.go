package database

import (
	"entgo.io/ent/dialect"
	"fmt"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/config"
	"github.com/janghanul090801/go-backend-clean-architecture-fiber/ent"
	_ "github.com/lib/pq"
)

// New returns data source name
func New() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.E.DBHost,
		config.E.DBPort,
		config.E.DBUser,
		config.E.DBPass,
		config.E.DBName,
	)

	return dsn
}

// NewClient returns an orm client
func NewClient() (*ent.Client, error) {
	var entOptions []ent.Option
	entOptions = append(entOptions, ent.Debug())

	dsn := New()

	return ent.Open(dialect.Postgres, dsn, entOptions...)
}
