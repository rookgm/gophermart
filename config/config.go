package config

import (
	"flag"
	"os"
	"sync"
)

const (
	defaultGMartServerAddress = ":8080"
	defaultGMartDatabaseDSN   = ""
	defaultAccrualSystemAddr  = ":8181"
	defaultLogLevel           = "debug"
)

type Config struct {
	GMartServerAddr   string
	GMartDatabaseDSN  string
	AccrualSystemAddr string
	LogLevel          string
}

var (
	once      sync.Once
	singleton *Config
)

// New returns new Config. It parses command line and environment variables only once.
func New() (*Config, error) {
	once.Do(func() {
		cfg := Config{}

		// initialize flags
		flag.StringVar(&cfg.GMartServerAddr, "a", defaultGMartServerAddress, "gopher mart server address")
		flag.StringVar(&cfg.GMartDatabaseDSN, "d", defaultGMartDatabaseDSN, "gopher mart database DSN")
		flag.StringVar(&cfg.AccrualSystemAddr, "r", defaultAccrualSystemAddr, "accrual system address")
		flag.StringVar(&cfg.LogLevel, "l", defaultLogLevel, "log level")

		flag.Parse()

		// if environment variable is set, then using it
		if runAddrEnv := os.Getenv("RUN_ADDRESS"); runAddrEnv != "" {
			cfg.GMartServerAddr = runAddrEnv
		}
		if dataBaseURIEnv := os.Getenv("DATABASE_URI"); dataBaseURIEnv != "" {
			cfg.GMartDatabaseDSN = dataBaseURIEnv
		}
		if accrualSysAddrEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualSysAddrEnv != "" {
			cfg.AccrualSystemAddr = accrualSysAddrEnv
		}
		if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
			cfg.LogLevel = logLevelEnv
		}

		singleton = &cfg
	})

	return singleton, nil
}
