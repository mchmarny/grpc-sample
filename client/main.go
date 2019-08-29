package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"os"
	"time"

	ptypes "github.com/golang/protobuf/ptypes"
	pb "github.com/mchmarny/grpc-sample/pkg/api/v1"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	logger     = log.New(os.Stdout, "", 0)
	serverAddr = flag.String("server", "", "Server address (host:port)")
	serverHost = flag.String("server-host", "", "Host name to which server IP should resolve")
	insecure   = flag.Bool("insecure", false, "Skip SSL validation? [false]")
	skipVerify = flag.Bool("skip-verify", false, "Skip server hostname verification in SSL validation [false]")
	streamMsgs = flag.Int("stream", 0, "Number of messages to stream [0]")
	author     = flag.String("author", "Sample Client", "The author of the content sent to server")
	message    = flag.String("message", "Hi there", "The body of the content sent to server")
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *serverHost != "" {
		opts = append(opts, grpc.WithAuthority(*serverHost))
	}
	if *insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		cred := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: *skipVerify,
		})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		logger.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewMessageServiceClient(conn)

	if *streamMsgs == 0 {
		send(client)
	} else {
		sendStream(client)
	}
}

func send(client pb.MessageServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	c := getContent()
	rep, err := client.Send(ctx, &pb.Request{Content: c})
	if err != nil {
		logger.Fatalf("Error while executing Send: %v", err)
	}
	logger.Printf("Unary Request/Unary Response\n Sent:\n  %+v\n Response:\n  %+v", c, rep)
}

func sendStream(client pb.MessageServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stream, err := client.SendStream(ctx)
	if err != nil {
		logger.Fatalf("Error client (%v) PingStream: %v", client, err)
	}

	waitCh := make(chan struct{})
	go func() {
		logger.Println("Unary Request/Stream Response")
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitCh)
				return
			}
			if err != nil {
				logger.Fatalf("Failed to receive a response: %v", err)
			}
			logger.Printf("  Stream[%d] - Server time: %s",
				in.GetIndex(), ptypes.TimestampString(in.GetReceivedOn()))
		}
	}()

	i := 0
	for i < *streamMsgs {
		if err := stream.Send(&pb.Request{Content: getContent()}); err != nil {
			logger.Fatalf("Failed to Send: %v", err)
		}
		i++
	}
	stream.CloseSend()
	<-waitCh
}

func getContent() *pb.Content {
	return &pb.Content{
		Body:      *message,
		Author:    *author,
		CreatedOn: ptypes.TimestampNow(),
	}
}
