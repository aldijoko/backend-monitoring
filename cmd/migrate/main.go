package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"monitoring-cctv-be/config"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.SSLMode,
	)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatalf("migrate init failed: %v", err)
	}

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	switch direction {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	default:
		log.Fatalf("unknown direction: %s (use up|down)", direction)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration completed")
}
