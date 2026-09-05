package main

import (
	"log"

	"monitoring-cctv-be/config"
	"monitoring-cctv-be/internal/repository"
	"monitoring-cctv-be/internal/service"
	"monitoring-cctv-be/pkg/database"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(&cfg.Database)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	users := repository.NewUserRepository(db)
	if err := service.EnsureBootstrapAdmin(users, cfg.AdminUsername, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("bootstrap admin failed: %v", err)
	}

	log.Printf("bootstrap admin ready: username=%s", cfg.AdminUsername)
}
