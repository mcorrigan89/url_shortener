package main

import (
	"os"
	"sync"

	"github.com/go-playground/form/v4"
	"github.com/mcorrigan89/url_shortener/internal/config"
	"github.com/mcorrigan89/url_shortener/internal/repositories"
	"github.com/mcorrigan89/url_shortener/internal/services"

	"github.com/rs/zerolog"
)

type application struct {
	config      *config.Config
	wg          *sync.WaitGroup
	logger      *zerolog.Logger
	services    *services.Services
	formDecoder *form.Decoder
}

func main() {

	cfg := config.Config{}

	config.LoadConfig(&cfg)

	logger := getLogger(cfg)

	db, err := openDBPool(cfg, &logger)
	if err != nil {
		logger.Err(err).Msg("Error opening database connection")
		os.Exit(1)
	}
	defer db.Close()

	wg := sync.WaitGroup{}

	repositories := repositories.NewRepositories(db, &cfg, &logger, &wg)
	services := services.NewServices(&repositories, &cfg, &logger, &wg)

	formDecoder := form.NewDecoder()

	app := &application{
		wg:          &wg,
		config:      &cfg,
		logger:      &logger,
		services:    &services,
		formDecoder: formDecoder,
	}

	err = app.serve()
	if err != nil {
		logger.Err(err).Msg("Error starting server")
		os.Exit(1)
	}
}
