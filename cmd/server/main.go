//go:generate protoc -I ../../pkg/helloworld --go_out=plugins=grpc:../../pkg/helloworld ../../pkg/helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/mkmik/k8s-grpc-tls-demo/pkg/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var (
	addr     = flag.String("listen", ":50052", "Listening address")
	certPath = flag.String("cert", "", "Path to TLS certificate")
	keyPath  = flag.String("key", "", "Path to TLS key")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
	hostname string
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: fmt.Sprintf("%q says: Hello %s", s.hostname, in.GetName())}, nil
}

func run(addr, certPath, keyPath string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	var opts []grpc.ServerOption
	if certPath != "" {
		creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(creds))
	}
	s := grpc.NewServer(opts...)

	reflection.Register(s)
	service.RegisterChannelzServiceToServer(s)

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	pb.RegisterGreeterServer(s, &server{hostname: hostname})

	log.Printf("listening on %q", addr)
	return s.Serve(lis)
}

func main() {
	flag.Parse()

	if err := run(*addr, *certPath, *keyPath); err != nil {
		log.Fatal(err)
	}
}
