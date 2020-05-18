package database

import (
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //Register postgres database/sql driver.
)

// Config required to open a database connection
type Config struct {
	Host       string
	Name       string
	User       string
	Password   string
	DisableTLS bool
}

// Open knows how to open a database connection
func Open(cfg Config) (*sqlx.DB, error) {
	q := url.Values{}

	q.Set("sslmode", "require")
	if cfg.DisableTLS {
		q.Set("sslmode", "disable")
	}
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}
