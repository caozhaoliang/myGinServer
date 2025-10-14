package server

import (
	"context"
	"fmt"
	"log"
	pb "myGinServer/utils/grpc-helloworld/protogen"
	"testing"
	"time"

	"net"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

const (
	port = ":50051"
)

// SayHello 实现简单的 RPC
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// SayHelloStream 实现服务端流式 RPC
func (s *server) SayHelloStream(in *pb.HelloRequest, stream pb.Greeter_SayHelloStreamServer) error {
	log.Printf("Stream request received: %v", in.GetName())

	for i := 0; i < 5; i++ {
		message := &pb.HelloReply{
			Message: fmt.Sprintf("Hello %s - message #%d", in.GetName(), i+1),
		}
		if err := stream.Send(message); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func TestServer(t *testing.T) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
