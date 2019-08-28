package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

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
	client := pb.NewPingServiceClient(conn)
	ping(client, "hello")
	pingStream(client, "hello")
}

func ping(client pb.PingServiceClient, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	rep, err := client.Ping(ctx, &pb.Request{Msg: msg})
	if err != nil {
		logger.Fatalf("%v.Ping failed %v: ", client, err)
	}
	logger.Printf("[unary-unary] ping response: %v", rep.GetMsg())
}

func pingStream(client pb.PingServiceClient, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.PingStream(ctx)
	if err != nil {
		logger.Fatalf("%v.(_) = _, %v", client, err)
	}

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				logger.Fatalf("Failed to receive a response: %v", err)
			}
			logger.Printf("[unary-stream] ping response: %s", in.GetMsg())
		}
	}()

	i := 0
	for i < *streamMsgs {
		if err := stream.Send(&pb.Request{Msg: fmt.Sprintf("%s-%d", msg, i)}); err != nil {
			logger.Fatalf("Failed to send a ping: %v", err)
		}
		i++
	}
	stream.CloseSend()
	<-waitc
}
