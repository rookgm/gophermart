package main

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/gophermart/config"
	"github.com/rookgm/gophermart/internal/accrual"
	"github.com/rookgm/gophermart/internal/auth"
	handler "github.com/rookgm/gophermart/internal/handler/http"
	"github.com/rookgm/gophermart/internal/logger"
	"github.com/rookgm/gophermart/internal/middleware"
	"github.com/rookgm/gophermart/internal/repository"
	"github.com/rookgm/gophermart/internal/repository/postgres"
	"github.com/rookgm/gophermart/internal/service"
	"github.com/rookgm/gophermart/internal/worker"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const authTokenKey = "f53ac685bbceebd75043e6be2e06ee07"
const shutdownTimeout = 5 * time.Second

func main() {

	// create new config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// initialize logger
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	// create context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// initialize database
	db, err := postgres.New(ctx, cfg.GMartDatabaseDSN)
	if err != nil {
		logger.Log.Fatal("Error initializing database", zap.Error(err))
	}
	defer db.Close()

	// migrate database
	err = db.Migrate()
	if err != nil {
		logger.Log.Fatal("Error migrating database", zap.Error(err))
	}

	tokenKey, err := hex.DecodeString(authTokenKey)
	if err != nil {
		logger.Log.Fatal("Error extracting token key", zap.Error(err))
	}
	token := auth.NewAuthToken(tokenKey)

	// dependency injection
	// accrual
	accrualHandler := accrual.NewAccrualClient(cfg.AccrualSystemAddr)

	// user
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, token)

	// auth
	authService := service.NewAuthService(userRepo, token)
	authHandler := handler.NewAuthHandler(authService)

	// order
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, accrualHandler)
	orderHandler := handler.NewOrderHandler(orderService)

	// balance
	balanceRepo := repository.NewBalanceRepository(db)
	balanceService := service.NewBalanceService(balanceRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	// order processor for accrual
	orderProc := worker.NewOrderProcessor(orderService)

	router := chi.NewRouter()

	router.Use(middleware.Logging(logger.Log))

	router.Post("/api/user/register", userHandler.RegisterUser())
	router.Post("/api/user/login", authHandler.LoginUser())

	// routes that require authentication
	router.Group(func(group chi.Router) {
		group.Use(handler.AuthMiddleware(token))
		group.Post("/api/user/orders", orderHandler.UploadUserOrder())
		group.Get("/api/user/orders", orderHandler.ListUserOrders())
		group.Get("/api/user/balance", balanceHandler.GetUserBalance())
		group.Post("/api/user/balance/withdraw", balanceHandler.UserBalanceWithdrawal())
		group.Get("/api/user/withdrawals", balanceHandler.GetUserWithdrawals())
	})

	// set server parameters
	srv := http.Server{
		Addr:    cfg.GMartServerAddr,
		Handler: router,
	}

	logger.Log.Info("Starting order processor")
	go func() {
		orderProc.ProcessOrders(ctx)
	}()

	logger.Log.Info("Starting server...")
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Error starting server", zap.Error(err))
		}
	}()

	logger.Log.Info("Server is started", zap.String("addr", cfg.GMartServerAddr))

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Error shutdown server", zap.Error(err))
	}

	logger.Log.Info("server is finished")
}
