//Package cmd is responsible for initializing grpc and http servers.
package cmd

import (
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
	baseServer "github.com/4m3ndy/amazon-prime-scrapper/pkg/server"
)

type ServerConfig struct {
	ServiceHTTPPort string `env:"AMAZON_PRIME_SCRAPPER_HTTP_PORT,required"`
}

// parseDBConfigurationFromEnv parses the configuration from the environment
func parseConfigurationFromEnv() (config, error) {
	cfg := config{}
	err := env.Parse(&cfg)
	return cfg, err
}

// RunServer ...
// start the health and grpc server
func RunServer() {

	logger.CreateLogger()
	// defer logger.InitializePanicHandler(true)

	cfg, err := parseConfigurationFromEnv()
	if err != nil {
		logger.Log().WithError(err).Panic("error parsing config")
	}

	server := baseServer.Server{HTTPPort: cfg.ServiceHTTPPort}
	server.RunServer(_registerHTTPServer)
}
