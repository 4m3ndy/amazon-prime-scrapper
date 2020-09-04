//
// Main file for microservice.
//
package main

import "github.com/4m3ndy/amazon-prime-scrapper/pkg/cmd"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cmd.RunServer(ctx)
}
