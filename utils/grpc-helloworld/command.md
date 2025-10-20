## protoc编译proto文件，生成对应的go文件
```
protoc --go_out=. --go-grpc_out=. proto/helloworld.proto
```

## 通过grpcurl命令行调用grpc服务进行调试
### 前置条件
1. 安装grpcurl
`go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`
2. 服务端需要注册反射 服务

```go
func main() {
    s := grpc.NewServer()
    pb.RegisterYourOwnServer(s, &server{})

    // Register reflection service on gRPC server.
    reflection.Register(s)

    s.Serve(lis)
}
```
3. 调用
```shell
1.查看服务端服务列表
  grpcurl -plaintext localhost:50051 list
2.查看服务端服务中的方法列表
  grpcurl -plaintext localhost:50051 list helloworld.Greeter
3.查看详细方法信息
  grpcurl -plaintext localhost:50051 describe helloworld.Greeter
4.查看请求体的详细内容
  grpcurl -plaintext localhost:50051 describe helloworld.HelloRequest

5.调用服务
  grpcurl -plaintext -d '{"name":"zhangSan"}' localhost:50051 helloworld.Greeter.SayHello
```