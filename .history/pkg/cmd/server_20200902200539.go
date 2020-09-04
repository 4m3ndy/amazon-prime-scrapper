//Package cmd is responsible for initializing grpc and http servers.
package cmd

import (
	"context"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/4m3ndy/amazon-prime-scrapper/logger"
	"github.com/4m3ndy/amazon-prime-scrapper/pkg/server"
)

func registerGRPCServerWithArgs(db *sqlx.DB, pubsubClient *pubsub.Client, environmentName string, xsdHandler *xsdvalidate.XsdHandler) func(*grpc.Server) {
	return func(grpcSrv *grpc.Server) {
		serviceAPI := api.NewSourceBranchServiceAPI(db, pubsubClient, environmentName, xsdHandler)
		eventsourcing.RegisterEventSourcingServiceServer(grpcSrv, serviceAPI)
		proto.RegisterSourceBranchServiceServer(grpcSrv, serviceAPI)
	}
}

func _registerHTTPServer(s *http.ServeMux) {
	// nothing to do
}

type config struct {
	ServiceGrpcPort string `env:"SOURCE_BRANCH_SERVICE_GRPC_PORT,required"`
	ServiceHTTPPort string `env:"SOURCE_BRANCH_SERVICE_HTTP_PORT,required"`

	ECDSAPublicKey string `env:"JWT_PUBLIC_KEY,required"`
	baseServer.ServerEnvConfig

	PubSubServiceAccountCredentials string `env:"PUBSUB_SERVICE_ACCOUNT_CREDENTIALS,required"`

	FeatureXMLSchemaURL string `env:"FEATURE_XML_SCHEMA_URL" envDefault:"/etc/featurexml/feature-xml-schema.xsd"`
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

	// Parse the ecdsa public key from the read key
	ecdsaPublicKey, err := jwt.ParseECPublicKeyFromPEM([]byte(cfg.ECDSAPublicKey))
	if err != nil {
		logger.Log().WithError(err).Panic("error parsing ecdsa public key file")
	}

	dbCfg, err := db.ParseConfigurationFromEnv()
	if err != nil {
		logger.Log().WithError(err).Panic("error parsing db config")
	}

	dataBase, err := db.CreateConnection(dbCfg)
	if err != nil {
		logger.Log().WithError(err).Panic("error connecting to database")
	}

	pubsubClient, err := createPubSubClient(cfg)
	if err != nil {
		logger.Log().WithError(err).Panic("error connecting to pubsub")
	}

	dbx := sqlx.NewDb(dataBase, "mysql")

	err = db.MigrateUp(dataBase)
	if err != nil {
		logger.Log().WithError(err).Panic("error migrating database")
	}

	err = xsdvalidate.Init()
	if err != nil {
		logger.Log().WithError(err).Panic("error initializing schema validation")
	}
	defer xsdvalidate.Cleanup()
	xsdHandler, err := xsdvalidate.NewXsdHandlerUrl(cfg.FeatureXMLSchemaURL, xsdvalidate.ParsErrDefault)
	if err != nil {
		logger.Log().WithError(err).Panic("error xml schema not found")
	}
	defer xsdHandler.Free()

	serverConfig := baseServer.ServerConfig{EcdsaPublicKey: ecdsaPublicKey, ServerEnvConfig: cfg.ServerEnvConfig}
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
