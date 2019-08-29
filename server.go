package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	ev "github.com/mchmarny/gcputil/env"
	ping "github.com/mchmarny/grpc-sample/pkg/api/v1"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	logger = log.New(os.Stdout, "", 0)
	port   = ev.MustGetEnvVar("PORT", "8080")
)

type pingServer struct {
}

func (p *pingServer) Ping(ctx context.Context, req *ping.Request) (*ping.Response, error) {
	return &ping.Response{Msg: fmt.Sprintf("%s - pong", req.Msg)}, nil
}

func (p *pingServer) PingStream(stream ping.PingService_PingStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			logger.Println("Client disconnected")
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "Failed to receive ping")
		}

		now := time.Now()
		logger.Printf("Replying to ping %s at %s\n", req.Msg, now)
		err = stream.Send(&ping.Response{
			Msg: fmt.Sprintf("pong at %s", now),
		})

		if err != nil {
			return errors.Wrap(err, "Failed to send pong")
		}
	}
}

func main() {

	hostPort := net.JoinHostPort("0.0.0.0", port)
	lis, err := net.Listen("tcp", hostPort)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", hostPort, err)
	}

	pingServer := &pingServer{}
	grpcServer := grpc.NewServer()
	ping.RegisterPingServiceServer(grpcServer, pingServer)

	grpcServer.Serve(lis)
	if err != grpcServer.Serve(lis) {
		logger.Fatalf("Failed while serving: %v", err)
	}
}
