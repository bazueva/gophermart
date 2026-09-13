package main

import (
	"flag"
	"os"

	configpkg "github.com/bazueva/gofermart/cmd/config"
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type config struct {
	ServerAddr           configpkg.ServerAddr `env:"RUN_ADDRESS"`
	DatabaseDSN          string               `env:"DATABASE_URI"`
	SecretKey            string               `env:"SECRET_KEY"`
	AccrualSystemAddress string               `env:"ACCRUAL_SYSTEM_ADDRESS"`

	logger *zap.Logger
}

func readConfig() (config, error) {
	cfg := config{
		ServerAddr: configpkg.ServerAddr{
			Host: "localhost",
			Port: 8080,
		},
		// падают тесты, поэтому добавлено дефолтное значение
		SecretKey: "K7#mP2!xQ9@vL4$z",
	}

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}

	err = parseFlags(&cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

func parseFlags(config *config) error {
	serverFlags := flag.NewFlagSet("", flag.ContinueOnError)
	serverFlags.Var(&config.ServerAddr, "a", "address http server")

	databaseDSN := serverFlags.String("d", config.DatabaseDSN, "Database DSN")
	secretKey := serverFlags.String("s", config.SecretKey, "Secret Key")
	accrualSystemAddress := serverFlags.String("r", config.AccrualSystemAddress, "accrual system address")

	if len(os.Args) > 1 {
		err := serverFlags.Parse(os.Args[1:])
		if err != nil {
			return err
		}
	}

	config.DatabaseDSN = *databaseDSN
	config.SecretKey = *secretKey
	config.AccrualSystemAddress = *accrualSystemAddress

	return nil
}
