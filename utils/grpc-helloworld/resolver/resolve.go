package resolver

import (
	"log"

	"google.golang.org/grpc/resolver"
)

const (
	myScheme      = "example"
	myServiceName = "my-custom-service:1234"
	backendAddr1  = "127.0.0.1:1"
	backendAddr2  = "127.0.0.1:50051"
	backendAddr3  = "127.0.0.1:50050"
)

type myResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

type myResolverBuilder struct{}

func (*myResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r := &myResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			myServiceName: {backendAddr1, backendAddr2, backendAddr3},
		},
	}
	r.start()
	return r, nil
}

func (*myResolverBuilder) Scheme() string { return myScheme }

func (r *myResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		log.Fatalf("UpdateState failed: %v", err)
	}
}

func (r *myResolver) ResolveNow(resolver.ResolveNowOptions) {}
func (r *myResolver) Close()                                {}

func init() {
	resolver.Register(&myResolverBuilder{})
}
