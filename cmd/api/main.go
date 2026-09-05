package main

import (
	"log"
	"time"

	"monitoring-cctv-be/config"
	"monitoring-cctv-be/internal/repository"
	"monitoring-cctv-be/internal/router"
	"monitoring-cctv-be/internal/service"
	"monitoring-cctv-be/pkg/database"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(&cfg.Database)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	media := service.NewMediaMTXService(cfg.MediaMTXAPIURL, cfg.MediaMTXHLSBaseURL)

	cameraRepo := repository.NewCameraRepository(db)
	edgeRepo := repository.NewEdgeRepository(db)

	cameraService := service.NewCameraService(cameraRepo, edgeRepo, media)
	cameraService.ReconcileMediaPaths()

	poller := service.NewCameraStatusPoller(cameraRepo, edgeRepo, media, 10*time.Second)
	stop := make(chan struct{})
	go poller.Run(stop)
	defer close(stop)

	r := router.Setup(db, cfg, media)

	log.Printf("listening on :%s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
