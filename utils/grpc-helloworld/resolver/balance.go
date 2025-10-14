package resolver

//
//import (
//	"context"
//	"fmt"
//	"hash/fnv"
//	"log"
//	"sync"
//
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/balancer"
//	"google.golang.org/grpc/balancer/base"
//	"google.golang.org/grpc/metadata"
//	"google.golang.org/grpc/resolver"
//)
//
//const (
//	HashPolicyName = "user_hash"
//)
//
//// NewHashBalancerBuilder 创建一个新的哈希负载均衡器构建器
//func NewHashBalancerBuilder() balancer.Builder {
//	return base.NewBalancerBuilder(
//		HashPolicyName,
//		&hashPickerBuilder{},
//		base.Config{HealthCheck: false},
//	)
//}
//
//// hashPickerBuilder 用于创建hashPicker
//type hashPickerBuilder struct{}
//
//// Build 根据当前可用的子连接创建一个新的picker
//func (b *hashPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
//	if len(info.ReadySCs) == 0 {
//		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
//	}
//
//	// 收集所有可用的子连接
//	var subConns []balancer.SubConn
//	for sc := range info.ReadySCs {
//		subConns = append(subConns, sc)
//	}
//
//	return &hashPicker{
//		subConns: subConns,
//	}
//}
//
//// hashPicker 实现了balancer.Picker接口，基于userId进行哈希选择
//type hashPicker struct {
//	subConns []balancer.SubConn
//	mu       sync.RWMutex
//}
//
//// Pick 选择一个子连接用于处理当前请求
//func (p *hashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
//	p.mu.RLock()
//	defer p.mu.RUnlock()
//
//	if len(p.subConns) == 0 {
//		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
//	}
//
//	// 从请求中获取userId
//	var userId string
//
//	// 方法2: 从metadata中获取userId (更通用的方式)
//	md, ok := metadata.FromOutgoingContext(info.Ctx)
//	if ok {
//		userIds := md.Get("user-id")
//		if len(userIds) > 0 {
//			userId = userIds[0]
//		}
//	}
//
//	// 如果没有获取到userId，使用一个默认值（可以根据需要修改策略）
//	if userId == "" {
//		userId = "default"
//	}
//
//	// 计算userId的哈希值
//	hash := fnv.New32a()
//	hash.Write([]byte(userId))
//	hashValue := hash.Sum32()
//
//	// 根据哈希值选择子连接
//	index := int(hashValue) % len(p.subConns)
//	subConn := p.subConns[index]
//	fmt.Println("Selected subConn:", index, subConn)
//	return balancer.PickResult{SubConn: subConn}, nil
//}
//
//// 为了能从上下文中获取请求，需要实现一个客户端拦截器
//func unaryClientInterceptor() grpc.UnaryClientInterceptor {
//	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
//		// 将请求存入上下文，以便在Picker中获取
//		newCtx := context.WithValue(ctx, "request", req)
//		return invoker(newCtx, method, req, reply, cc, opts...)
//	}
//}
//
//// hashBalancer 实现 balancer.Balancer 接口
//type hashBalancer struct {
//	cc       balancer.ClientConn         // 客户端连接
//	subConns map[string]balancer.SubConn // 用地址作为 key 存储所有子连接
//}
//
//func newHashBalancer(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
//	return &hashBalancer{
//		cc:       cc,
//		subConns: make(map[string]balancer.SubConn), // 初始化子连接映射
//	}
//}
//func (b *hashBalancer) ExitIdle() {
//	fmt.Println("exitIdle", b.subConns)
//}
//func (b *hashBalancer) ResolverError(err error) {
//	fmt.Println(err)
//}
//func (b *hashBalancer) UpdateSubConnState(sc balancer.SubConn, state balancer.SubConnState) {
//	fmt.Println(sc, state)
//}
//
//// UpdateClientConnState 处理客户端连接状态更新（核心方法）
//func (b *hashBalancer) UpdateClientConnState(state balancer.ClientConnState) error {
//	// 1. 从状态中获取最新的地址列表
//	addresses := state.ResolverState.Addresses
//	if len(addresses) == 0 {
//		return fmt.Errorf("没有可用的服务地址")
//	}
//
//	// 2. 为新地址创建子连接（如果不存在）
//	for _, addr := range addresses {
//		addrStr := addr.Addr // 用地址字符串作为唯一标识
//		if _, exists := b.subConns[addrStr]; !exists {
//			// 为该地址创建新的子连接
//			sc, err := b.cc.NewSubConn([]resolver.Address{addr}, balancer.NewSubConnOptions{})
//			if err != nil {
//				log.Printf("创建子连接失败（%s）：%v", addrStr, err)
//				continue
//			}
//			b.subConns[addrStr] = sc // 存入映射
//			log.Printf("已添加新的服务实例：%s", addrStr)
//		}
//	}
//
//	// 3. 移除已不存在的地址（可选，处理服务下线）
//	for addrStr := range b.subConns {
//		found := false
//		for _, addr := range addresses {
//			if addr.Addr == addrStr {
//				found = true
//				break
//			}
//		}
//		if !found {
//			b.cc.RemoveSubConn(b.subConns[addrStr]) // 移除子连接
//			delete(b.subConns, addrStr)
//			log.Printf("已移除服务实例：%s", addrStr)
//		}
//	}
//
//	// 4. 打印当前就绪的服务实例数量
//	log.Printf("当前可用服务实例总数：%d", len(b.subConns))
//	return nil
//}
//
//// 其他接口方法（空实现，按需补充）
//func (b *hashBalancer) Close() {}
