package main

import (
	"io"
	"log"
	"net"
	"os"

	ptypes "github.com/golang/protobuf/ptypes"
	ev "github.com/mchmarny/gcputil/env"
	pb "github.com/mchmarny/grpc-sample/pkg/api/v1"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	logger = log.New(os.Stdout, "", 0)
	port   = ev.MustGetEnvVar("PORT", "8080")
)

type messageService struct {
}

func (p *messageService) Send(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	return &pb.Response{
		Index:      1,
		ReceivedOn: ptypes.TimestampNow(),
		Content:    req.GetContent(),
	}, nil
}

func (p *messageService) SendStream(stream pb.MessageService_SendStreamServer) error {
	var i int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			logger.Println("Client disconnected")
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "Failed to receive ping")
		}

		c := req.GetContent()
		i++ // TODO: clean this up
		logger.Printf("Replying to send[%d]: %+v", i, c)

		err = stream.Send(&pb.Response{
			Index:      i,
			ReceivedOn: ptypes.TimestampNow(),
			Content:    req.GetContent(),
		})

		if err != nil {
			return errors.Wrap(err, "Failed to send pong")
		}
	}
}

func main() {

	hostPort := net.JoinHostPort("0.0.0.0", port)
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", hostPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, &messageService{})

	if err != grpcServer.Serve(listener) {
		logger.Fatalf("Failed while serving: %v", err)
	}
}
