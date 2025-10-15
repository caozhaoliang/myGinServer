package resolver

import (
	"log"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

func init() {
	// 注册带权重的解析器
	resolver.Register(&weightedResolverBuilder{})
}

type AddrMetadata struct {
	Weight int
}

// 权重定义
var weightMap = map[string]int{
	backendAddr1: 0, //权重定为1
	backendAddr2: 3, //权重定为3
	backendAddr3: 1,
}

type weightedResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

type weightedResolverBuilder struct{}

// RegisterResolver 注册加权解析器
func RegisterResolver() {
	resolver.Register(&weightedResolverBuilder{})
}

func (*weightedResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r := &weightedResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			myServiceName: {backendAddr1, backendAddr2, backendAddr3},
		},
	}
	r.start()
	return r, nil
}

func (*weightedResolverBuilder) Scheme() string { return myScheme }

func (r *weightedResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		// 获取权重
		weight := weightMap[s]
		// 创建权重元数据
		meta := AddrMetadata{Weight: weight}
		// 创建属性
		attrs := attributes.New(AddrMetadata{}, meta)
		// 创建地址+属性
		addrs[i] = resolver.Address{
			Addr:       s,
			Attributes: attrs,
		}
		log.Printf("加权解析器: 添加服务地址 %s 权重 %d, Attributes: %v", s, weight, attrs)
	}

	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		log.Fatalf("UpdateState失败: %v", err)
	}
}

func (r *weightedResolver) ResolveNow(resolver.ResolveNowOptions) {}
func (r *weightedResolver) Close()                                {}
