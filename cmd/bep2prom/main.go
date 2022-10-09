package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	bes "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/grpc"

	"github.com/ldx/bep2prom/pkg/metrics"
	"github.com/ldx/bep2prom/pkg/server"
)

type Options struct {
	ListenAddr string `short:"l" long:"listen-addr" description:"Start listener on this host:port" default:":8799"`
}

func main() {
	opts := Options{}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	grpcServer := grpc.NewServer()
	bes.RegisterPublishBuildEventServer(grpcServer, server.New())
	go func() {
		err = metrics.Serve()
		if err != nil {
			log.Fatalf("Failed to serve metrics: %v", err)
		}
	}()
	go func() {
		log.Printf("Listening on %s", opts.ListenAddr)
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	for {
		select {
		case s := <-sigs:
			log.Printf("Received signal %v, exiting", s)
			grpcServer.GracefulStop()
			os.Exit(0)
		}
	}
}
