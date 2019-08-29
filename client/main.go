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
	streamMsgs = flag.Int("stream-msg-num", 10, "Number of stream messages [10]")
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
	send(client)
	sendStream(client)
}

func send(client pb.MessageServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	rep, err := client.Send(ctx, &pb.Request{Content: getContent()})
	if err != nil {
		logger.Fatalf("Error while executing Send: %v", err)
	}
	logger.Printf("unary request, unary response\n  Sent[%d]: %+v",
		rep.GetIndex(), rep.GetReceivedOn())
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
		logger.Println("unary request, stream response")
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitCh)
				return
			}
			if err != nil {
				logger.Fatalf("Failed to receive a response: %v", err)
			}
			logger.Printf("  Sent[%d]: %+v",
				in.GetIndex(), in.GetReceivedOn())
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
