package main

import (
	"log"

	"starter-backend/internal/config"
	"starter-backend/internal/database"
	"starter-backend/internal/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.DatabaseDSN, cfg.AppEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}
	packages, err := seed.DemoPackages(db, cfg.AppEnv)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range packages {
		log.Printf("demo package ready: id=%d name=%q", item.ID, item.Name)
	}
}
