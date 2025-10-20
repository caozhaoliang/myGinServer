package main

import (
	"context"
	"fmt"
	"log"
	pb "myGinServer/utils/grpc-helloworld/protogen"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedGreeterServer
	addr string
}

const (
	port = ":50051"
)

// SayHello 实现简单的 RPC
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v from:%s", in.GetName(), s.addr)
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
func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	addrs := []string{"localhost:50050", "localhost:50051"}
	for _, addr := range addrs {
		go startServer(addr)
	}

	<-stop
	log.Println("Shutting down servers...")

}
func startServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{addr: addr})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
