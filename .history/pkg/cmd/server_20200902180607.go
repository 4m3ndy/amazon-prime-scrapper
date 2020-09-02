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
	"strings"

//	"github.com/4m3ndy/amazon-prime-scrapper/pkg/logger"
//	"github.com/4m3ndy/amazon-prime-scrapper/pkg/server"
)

type ServerConfig struct {
	HttpPort string `env:"AMAZON_PRIME_SCRAPPER_HTTP_PORT,required"`
}

// RunServer runs the http server
func RunServer() {
	serverConfig := getServerConfig()

	server.RunConfig(server.ServerConfig{
		HttpPort:           serverConfig.HttpPort,
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
