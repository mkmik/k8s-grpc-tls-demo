// Package main implements a client for Greeter service.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/mkmik/k8s-grpc-tls-demo/pkg/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	addr      = flag.String("addr", "", "Address of gRPC server (mandatory)")
	plaintext = flag.Bool("plaintext", false, "Use grpc without tls")
)

const (
	defaultName = "world"
)

func run(plaintext bool, address, name string) error {
	// Set up a connection to the server.
	config := &tls.Config{}
	transport := grpc.WithTransportCredentials(credentials.NewTLS(config))
	if plaintext {
		transport = grpc.WithInsecure()
	}
	conn, err := grpc.Dial(address, grpc.WithBlock(), grpc.WithBalancerName("round_robin"), transport)
	if err != nil {
		return fmt.Errorf("did not connect: %w", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		before := time.Now()
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s (talking to %q took %v)", r.GetMessage(), address, time.Since(before))
	}
	return nil
}

func main() {
	flag.Parse()

	if *addr == "" {
		flag.Usage()
		os.Exit(1)
	}

	name := defaultName
	if flag.NArg() > 1 {
		name = flag.Arg(1)
	}

	if err := run(*plaintext, *addr, name); err != nil {
		log.Fatal(err)
	}
}
