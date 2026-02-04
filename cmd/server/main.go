package main

import (
	"log"
	"net/http"

	httpapp "qr-doorbell/internal/api"
	"qr-doorbell/internal/config"
	"qr-doorbell/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("error cargando configuración:", err)
	}

	// DB opcional: si DATABASE_URL está vacía, se usa modo in-memory.
	var db *storage.DB
	if cfg.DatabaseURL != "" {
		db, err = storage.OpenPostgres("postgres", cfg.DatabaseURL)
		if err != nil {
			log.Fatal("error conectando a Postgres:", err)
		}
		log.Println("Conectado a Postgres")
	} else {
		log.Println("DATABASE_URL no configurada, usando almacenamiento en memoria")
	}

	router := httpapp.NewRouter(cfg, db)

	log.Println("Server listening on", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), router); err != nil {
		log.Fatal(err)
	}
}
