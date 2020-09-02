//
// Service setup and configuration.
//
package cmd

import (
	grpctrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/google.golang.org/grpc"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net"
	"os"

	"github.com/freiheit-com/venus/pkg/logger"
	"github.com/freiheit-com/venus/pkg/server"
	pb "github.com/freiheit-com/venus/services/hello-service/pkg/proto"
	"github.com/freiheit-com/venus/services/hello-service/pkg/service"
)

type ServerConfig struct {
	GrpcPort string `env:"TEMPLATE_SERVICE_GRPC_PORT,required"`
	HttpPort string `env:"TEMPLATE_SERVICE_HTTP_PORT,required"`
}

// RunServer runs the grpc and http servers
func RunServer() {
	serverConfig := getServerConfig()

	server.RunConfig(server.ServerConfig{
		GrpcPort:           serverConfig.GrpcPort,
		HttpPort:           serverConfig.HttpPort,
		RegisterGrpcServer: grpcSetup,
		GrpcServerOptions: []grpc.ServerOption{
			grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor()),
		},
	})
}

func grpcSetup(grpcSrv *grpc.Server) server.Shutdownable {
	tracer.Start(tracer.WithAgentAddr(net.JoinHostPort(os.Getenv("DD_AGENT_HOST"), "8126")))

	serviceS := service.NewService()
	pb.RegisterServiceServer(grpcSrv, serviceS)
	reflection.Register(grpcSrv)
	return server.ShutdownAll{serviceS, server.CancelFunc(tracer.Stop)}
}

func getServerConfig() ServerConfig {
	var serverConfig ServerConfig
	err := env.Parse(&serverConfig)
	if err != nil {
		logger.Log().WithError(err).Panic("could not load server config")
	}
	return serverConfig
}
