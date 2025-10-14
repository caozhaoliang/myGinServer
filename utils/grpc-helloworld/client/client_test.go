package client

import (
	"context"
	"log"
	"testing"
	"time"

	pb "myGinServer/utils/grpc-helloworld/protogen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051"
)

func TestClient(t *testing.T) {
	// 建立连接
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 调用简单 RPC
	name := "World"
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Simple RPC response: %s", r.GetMessage())
	// 调用流式 RPC
	log.Println("Starting stream RPC...")
	streamCtx, streamCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer streamCancel()

	stream, err := c.SayHelloStream(streamCtx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not start stream: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			log.Printf("Stream finished: %v", err)
			break
		}
		log.Printf("Stream response: %s", reply.GetMessage())
	}

}
