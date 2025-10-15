package client

import (
	"context"
	"fmt"
	"log"
	pb "myGinServer/utils/grpc-helloworld/protogen"
	_ "myGinServer/utils/grpc-helloworld/resolver"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	address1 = "localhost:50051"
	address2 = "localhost:50052"
	address  = "example:///my-custom-service:1234"
)

func TestClient(t *testing.T) { // 注册自定义负载均衡器
	// 建立连接
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//这里增加ound_robin
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// 设置超时上下文
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	// 调用简单 RPC
	//name := "World"
	//r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Simple RPC response: %s", r.GetMessage())
	// 调用流式 RPC
	//log.Println("Starting stream RPC...")
	//streamCtx, streamCancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer streamCancel()
	//
	//stream, err := c.SayHelloStream(streamCtx, &pb.HelloRequest{Name: name})
	//if err != nil {
	//	log.Fatalf("could not start stream: %v", err)
	//}
	//
	//for {
	//	reply, err := stream.Recv()
	//	if err != nil {
	//		log.Printf("Stream finished: %v", err)
	//		break
	//	}
	//	log.Printf("Stream response: %s", reply.GetMessage())
	//}

	// 测试不同userId的路由情况
	userIds := []string{"user1", "user2", "user3", "user1", "user2", "user3", "user4"}
	for _, uid := range userIds {
		md := metadata.Pairs("user-id", "123456")
		ctxMd := metadata.NewOutgoingContext(context.Background(), md)
		r, err := c.SayHello(ctxMd, &pb.HelloRequest{Name: uid})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Hash RPC response: %s", r.GetMessage())
	}
}

func TestWeightedClient(t *testing.T) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"weighted_round_robin"}`),
	)
	if err != nil {
		log.Fatalf("客户端: 连接失败: %v", err)
	}
	defer conn.Close()
	log.Println("客户端: 连接成功!")
	c := pb.NewGreeterClient(conn)

	for i := 1; i <= 10; i++ {
		md := metadata.Pairs("user-id", "123456")
		ctxMd := metadata.NewOutgoingContext(context.Background(), md)
		r, err := c.SayHello(ctxMd, &pb.HelloRequest{Name: fmt.Sprintf("user%d", i)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Hash RPC response: %s", r.GetMessage())
	}
}
