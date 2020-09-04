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
	cfg := ServerConfig{}
	err := env.Parse(&cfg)
	return cfg, err
}

func RunServer(ctx context.Context) (err error) {

	logger.CreateLogger()
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "okay")
		},
	))

	srv := &http.Server{
		Addr:    ":6969",
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	log.Printf("server started")

	<-ctx.Done()

	log.Printf("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed:%+s", err)
	}

	log.Printf("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	if err := serve(ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}
