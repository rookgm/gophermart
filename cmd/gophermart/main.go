package main

import (
	"context"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/gophermart/config"
	"github.com/rookgm/gophermart/internal/auth"
	handler "github.com/rookgm/gophermart/internal/handler/http"
	"github.com/rookgm/gophermart/internal/middleware"
	"github.com/rookgm/gophermart/internal/repository"
	"github.com/rookgm/gophermart/internal/repository/postgres"
	"github.com/rookgm/gophermart/internal/service"
	"go.uber.org/zap"
	"log"
	"net/http"
)

const authTokenKey = "f53ac685bbceebd75043e6be2e06ee07"

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

	// initialize database
	db, err := postgres.New(ctx, cfg.GMartDatabaseDSN)
	if err != nil {
		logger.Error("Error initializing database", zap.Error(err))
	}
	defer db.Close()

	// migrate database
	err = db.Migrate()
	if err != nil {
		logger.Fatal("Error migrating database", zap.Error(err))
	}

	tokenKey, err := hex.DecodeString(authTokenKey)
	if err != nil {
		logger.Fatal("Error extracting token key", zap.Error(err))
	}
	token := auth.NewAuthToken(tokenKey)

	// dependency injection
	// user
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, token)

	// auth
	authService := service.NewAuthService(userRepo, token)
	authHandler := handler.NewAuthHandler(authService)

	// order
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	// balance
	balanceRepo := repository.NewBalanceRepository(db)
	balanceService := service.NewBalanceService(balanceRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	router := chi.NewRouter()

	router.Use(middleware.Logging(logger))

	router.Post("/api/user/register", userHandler.RegisterUser())
	router.Post("/api/user/login", authHandler.LoginUser())

	// routes that require authentication
	router.Group(func(group chi.Router) {
		group.Use(middleware.Auth(token))
		group.Post("/api/user/orders", orderHandler.UploadUserOrder())
		group.Get("/api/user/orders", orderHandler.ListUserOrders())
		group.Get("/api/user/balance", balanceHandler.GetUserBalance())
		group.Post("/api/user/balance/withdraw", balanceHandler.UserBalanceWithdrawal())
		group.Get("/api/user/withdrawals", balanceHandler.GetUserWithdrawals())
	})

	logger.Info("Running server", zap.String("addr", cfg.GMartServerAddr))

	if err := http.ListenAndServe(cfg.GMartServerAddr, router); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}

	return
}
