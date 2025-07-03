package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rookgm/gophermart/config"
	"github.com/rookgm/gophermart/internal/gmart"
	"github.com/rookgm/gophermart/internal/handler"
	"github.com/rookgm/gophermart/internal/middleware"
	"go.uber.org/zap"
	"log"
	"net/http"
)

// newLogger creates logger with log level
func newLogger(level string) (*zap.Logger, error) {

	loggerLvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.Level = loggerLvl

	return loggerCfg.Build()
}

func main() {

	// create new config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// initialize logger
	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync()

	// create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// initialize database pool
	dbpool, err := pgxpool.New(ctx, cfg.GMartDatabaseDSN)
	if err != nil {
		logger.Error("Error initializing database pool", zap.Error(err))
	}
	defer dbpool.Close()

	// ping database pool
	if err := dbpool.Ping(ctx); err != nil {
		logger.Fatal("Error ping database pool", zap.Error(err))
	}

	// create repository
	repo := gmart.NewRepository(dbpool)

	// create service
	service := gmart.NewService(repo, cfg)

	// create handler
	handler := handler.New(service, logger)

	router := chi.NewRouter()

	mlogger := middleware.Logging(logger)
	router.Use(mlogger)

	router.Route("/", func(r chi.Router) {
		router.Post("/api/user/register", handler.RegisterUser())
		router.Post("/api/user/login", handler.AuthenticationUser())
	})

	logger.Info("Running server", zap.String("addr", cfg.GMartServerAddr))

	if err := http.ListenAndServe(cfg.GMartServerAddr, router); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}

	return
}
