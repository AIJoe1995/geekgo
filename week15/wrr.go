package main

import (
	"go.uber.org/atomic"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

const name = "custom_wrr"

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{}, base.Config{
		HealthCheck: false,
	}))
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for subConn, subConnInfo := range info.ReadySCs {
		cc := &conn{
			cc:        subConn,
			available: *atomic.NewBool(true),
		}
		md, ok := subConnInfo.Address.Metadata.(map[string]any)
		if ok {
			weightVal := md["weight"]
			weight, _ := weightVal.(float64)
			cc.weight = int(weight)
		}
		if cc.weight == 0 {
			cc.weight = 10
		}
		cc.currentWeight = cc.weight
		conns = append(conns, cc)
	}
	return &Picker{
		conns:     conns,
		max_limit: 1000,
		min_limit: 10,
	}

}

type conn struct {
	weight        int
	labels        []string
	currentWeight int
	cc            balancer.SubConn
	available     atomic.Bool
	// vip组 非vip组
	group string
	mu    sync.Mutex
}

type Picker struct {
	conns     []*conn
	mutex     sync.Mutex
	max_limit int
	min_limit int
	delta     int
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var total int
	var maxConn *conn
	for _, cc := range p.conns {
		if !cc.available.Load() {
			continue
		}
		total += cc.weight
		cc.currentWeight = cc.currentWeight + cc.weight
		if maxConn == nil || cc.currentWeight > maxConn.currentWeight {
			maxConn = cc
		}

	}
	maxConn.currentWeight = maxConn.currentWeight - total
	return balancer.PickResult{
		SubConn: maxConn.cc,
		Done: func(info balancer.DoneInfo) {
			err := info.Err
			if err == nil {
				// 增加weight
				maxConn.mu.Lock()
				maxConn.weight += p.delta
				if maxConn.weight > p.max_limit {
					maxConn.weight = p.max_limit
				}
				maxConn.mu.Unlock()
				return
			}

			maxConn.mu.Lock()
			maxConn.weight -= p.delta
			if maxConn.weight < p.min_limit {
				maxConn.weight = p.min_limit
			}
			maxConn.mu.Unlock()
		},
	}, nil
}

func (p *Picker) healthCheck(cc *conn) bool {
	// 调用 grpc 内置的那个 health check 接口
	return true
}
