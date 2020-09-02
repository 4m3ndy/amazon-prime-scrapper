//
// Service setup and configuration.
//
package cmd

import (
	"net"
	"os"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/freiheit-com/venus/pkg/logger"
	"github.com/freiheit-com/venus/pkg/server"
)

type ServerConfig struct {
	HttpPort string `env:"AMAZON_PRIME_SCRAPPER_HTTP_PORT,required"`
}

// RunServer runs the grpc and http servers
func RunServer() {
	serverConfig := getServerConfig()

	server.RunConfig(server.ServerConfig{
		HttpPort:           serverConfig.HttpPort,
		RegisterGrpcServer: grpcSetup,
		GrpcServerOptions: []grpc.ServerOption{
			grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor()),
		},
	})
}


func getServerConfig() ServerConfig {
	var serverConfig ServerConfig
	err := env.Parse(&serverConfig)
	if err != nil {
		logger.Log().WithError(err).Panic("could not load server config")
	}
	return serverConfig
}
