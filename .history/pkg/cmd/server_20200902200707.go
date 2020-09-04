//Package cmd is responsible for initializing grpc and http servers.
package cmd

import (
	"context"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/4m3ndy/amazon-prime-scrapper/logger"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/server"
)

func _registerHTTPServer(s *http.ServeMux) {
	// nothing to do
}

type config struct {
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
	defer logger.InitializePanicHandler(true)

	cfg, err := parseConfigurationFromEnv()
	if err != nil {
		logger.Log().WithError(err).Panic("error parsing config")
	}

	serverConfig := server.ServerConfig{ServerEnvConfig: cfg.ServerEnvConfig}
	server := baseServer.Server{GrpcPort: cfg.ServiceGrpcPort, HTTPPort: cfg.ServiceHTTPPort}
	server.RunServer(registerGRPCServerWithArgs(dbx, pubsubClient, cfg.EnvironmentName, xsdHandler), _registerHTTPServer, serverConfig)
}

func createPubSubClient(cfg config) (*pubsub.Client, error) {
	var options []option.ClientOption
	// only use service account credentials if they are set. This allows us to develop locally against pubsub.
	if cfg.PubSubServiceAccountCredentials != "" {
		options = append(options, option.WithCredentialsJSON([]byte(cfg.PubSubServiceAccountCredentials)))
	}
	return pubsub.NewClient(context.Background(), cfg.ProjectID, options...)
}
