package resolver

import (
	"log"
	"math"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const (
	Name          = "weighted_round_robin"
	DefaultWeight = 1
)

func init() {
	log.Println("WeightedRoundRobinLB: Registering balancer builder with name:", Name)
	balancer.Register(newBuilder())
}

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &wrrPickerBuilder{}, base.Config{HealthCheck: false})
}

type wrrPickerBuilder struct{}

func (pb *wrrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): Build called. Have %d Ready SubConns.", len(info.ReadySCs))

	if len(info.ReadySCs) == 0 {
		log.Println("WeightedRoundRobinLB (wrrPickerBuilder): No ready SubConns. Returning ErrNoSubConnAvailable.")
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var subConns []*weightedSubConn
	totalWeight := 0

	for sc, scInfo := range info.ReadySCs {
		weight := DefaultWeight
		var fetchedMetadata AddrMetadata
		var metadataOk bool
		if scInfo.Address.Attributes != nil {
			if val := scInfo.Address.Attributes.Value((AddrMetadata{})); val != nil {
				fetchedMetadata, metadataOk = val.(AddrMetadata)
			}
		}

		if metadataOk {
			log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): SubConn %s, successfully retrieved AddrMetadata: %+v", scInfo.Address.Addr, fetchedMetadata)
			if fetchedMetadata.Weight > 0 {
				weight = fetchedMetadata.Weight
				log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): SubConn %s, using metadata weight: %d", scInfo.Address.Addr, weight)
			} else {
				log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): SubConn %s, metadata weight is %d (<=0). Using default weight: %d", scInfo.Address.Addr, fetchedMetadata.Weight, DefaultWeight)
			}
		} else {
			// Log why metadata wasn't used
			if scInfo.Address.Attributes == nil {
				log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): SubConn %s, no attributes found. Using default weight: %d", scInfo.Address.Addr, DefaultWeight)
			} else {
				val := scInfo.Address.Attributes.Value(AddrMetadata{})
				if val == nil {
					log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): SubConn %s, attributes present, but key 'AddrMetadata{}' (or its equivalent) not found. Using default weight: %d", scInfo.Address.Addr, DefaultWeight)
				} else {
					log.Printf("WeightedRoundRobinLB (wrrPickerBuilder): SubConn %s, attributes present and key 'AddrMetadata{}' (or its equivalent) found, but value is of WRONG TYPE (got %T, expected AddrMetadata). Using default weight: %d", scInfo.Address.Addr, val, DefaultWeight)
				}
			}
		}

		subConns = append(subConns, &weightedSubConn{
			sc:            sc,
			address:       scInfo.Address.Addr,
			weight:        weight,
			currentWeight: 0, // Initialized to 0 as per smooth WRR
		})
		totalWeight += weight
	}

	return &wrrPicker{
		subConns:    subConns,
		totalWeight: totalWeight,
	}
}

// weightedSubConn holds a SubConn and its associated weight information for WRR.
type weightedSubConn struct {
	sc            balancer.SubConn
	address       string
	weight        int // Static weight assigned to this SubConn
	currentWeight int // Dynamic weight, updated during picking
}

// wrrPicker is a Picker that implements the weighted round-robin algorithm.
type wrrPicker struct {
	subConns    []*weightedSubConn
	mu          sync.Mutex
	totalWeight int
}

// Pick implements the Picker interface.
func (p *wrrPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.subConns) == 0 {
		// log.Println("WeightedRoundRobinLB (wrrPicker): Pick - No SubConns available. Returning ErrNoSubConnAvailable.")
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var bestSc *weightedSubConn
	maxCurrentWeight := math.MinInt

	for _, wsc := range p.subConns {
		wsc.currentWeight += wsc.weight
		if wsc.currentWeight > maxCurrentWeight {
			maxCurrentWeight = wsc.currentWeight
			bestSc = wsc
		}
	}

	if bestSc == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	bestSc.currentWeight -= p.totalWeight
	return balancer.PickResult{SubConn: bestSc.sc, Done: nil}, nil
}
