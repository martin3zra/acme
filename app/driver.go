package app

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/martin3zra/playsql"
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

	// Pilot: a playsql handle on the same database, used by the vendor read
	// path. Shares the DSN (a separate pool) until playsql can wrap *sql.DB.
	play, err := playsql.OpenDSN("postgres", s.resolveConnectionSrtring())
	if err != nil {
		panic(err)
	}
	s.play = play
}

func (s *Server) resolveConnectionSrtring() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s timezone=America/Santo_Domingo",
		s.config.db.host,
		s.config.db.port,
		s.config.db.name,
		s.config.db.username,
		s.config.db.password,
		s.config.db.sslmode,
	)
}
