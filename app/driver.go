package app

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func (s *Server) openDatabaseConnection() {
	db, err := sql.Open("postgres", s.resolveConnectionSrtring())
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(time.Millisecond * 300)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	s.db = db
}

func (s *Server) resolveConnectionSrtring() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		s.config.db.host,
		s.config.db.port,
		s.config.db.name,
		s.config.db.username,
		s.config.db.password,
	)
}
