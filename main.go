package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/NYTimes/gcs-helper/v3/handlers"
	"github.com/google/gops/agent"
	"google.golang.org/api/option"
	ghttp "google.golang.org/api/transport/http"
)

const version = "3.1.1"

func main() {
	handleFlags()
	err := agent.Listen(agent.Options{})
	if err != nil {
		log.Fatalf("could not start gops agent: %v", err)
	}
	defer agent.Close()
	config, err := handlers.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := config.Logger()
	hc, err := httpClient(config.Client)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize http client")
	}
	client, err := storage.NewClient(context.Background(), option.WithHTTPClient(hc))
	if err != nil {
		logger.WithError(err).Fatal("failed to create storage client instance")
	}
	handler := getHandler(config, client)
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		logger.WithField("listenAddr", config.Listen).WithError(err).Fatal("failed to start listener")
	}

	logger.Infof("Listening on %s...", listener.Addr())
	err = http.Serve(listener, handler)
	if err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
}

func httpClient(c handlers.ClientConfig) (*http.Client, error) {
	baseTransport := http.Transport{
		IdleConnTimeout: c.IdleConnTimeout,
		MaxIdleConns:    c.MaxIdleConns,
	}
	transport, err := ghttp.NewTransport(context.Background(), &baseTransport, option.WithScopes(storage.ScopeReadOnly))
	return &http.Client{
		Timeout:   c.Timeout,
		Transport: transport,
	}, err
}

func handleFlags() {
	printVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *printVersion {
		fmt.Printf("gcs-helper %s\n", version)
		os.Exit(0)
	}
}
