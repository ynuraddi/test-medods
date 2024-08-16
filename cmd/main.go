package main

import (
	"errors"
	"log"
	"medods/config"
	"medods/internal/repository"
	"medods/internal/service"
	router "medods/internal/transport/http"
	httpserver "medods/pkg/httpServer"
	"medods/pkg/logger"
	"medods/pkg/postgres"
	"medods/pkg/smtp"
	"net/http"
	"os"
	"os/signal"

	"github.com/golang-migrate/migrate/v4"
)

func main() {
	config, err := config.MustLoad()
	if err != nil {
		panic(err)
	}

	logger := logger.New(config.Log.Level, false)
	logger.Debug("Logger initializated")

	pg, err := postgres.New(&postgres.Config{
		DSN:          config.PG.DSN,
		MigrationURL: config.PG.MigrationURL,
	})
	if err != nil {
		panic(err)
	}

	if err := pg.MigrateUP(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}

	repo := repository.New(pg.Conn)

	smtp := smtp.New(&smtp.Config{
		Host:     config.SMTP.HOST,
		Port:     config.SMTP.PORT,
		Username: config.SMTP.USER,
		Password: config.SMTP.PASS,
		From:     config.SMTP.FROM,
	})

	service := service.New(config, repo, smtp, logger)

	router := router.NewRouter(service, logger)

	server := httpserver.New(router, &httpserver.Config{
		Port: config.HTTP.Port,
	})

	gracefullShutdown(func() {
		if err := server.Shutdown(); err != nil {
			log.Fatalf("Gracefull shutdown is failed: %s", err.Error())
		}
		pg.Close()
	})

	err = <-server.Notify()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server stopped with: %s", err.Error())
		return
	}
	logger.Info("Server stopped gracefully.")
}

func gracefullShutdown(shutdownFunc func()) {
	osC := make(chan os.Signal, 1)
	signal.Notify(osC, os.Interrupt)

	go func() {
		log.Println(<-osC)
		shutdownFunc()
	}()
}
