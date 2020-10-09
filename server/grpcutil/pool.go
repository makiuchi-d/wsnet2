package grpcutil

import (
	"sync"

	"google.golang.org/grpc"
)

type Pool struct {
	mu   sync.Mutex
	opts []grpc.DialOption
	pool map[string]*grpc.ClientConn
}

func NewPool(options ...grpc.DialOption) *Pool {
	return &Pool{
		opts: options,
		pool: make(map[string]*grpc.ClientConn),
	}
}

func (p *Pool) Get(target string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	cc, ok := p.pool[target]
	if ok {
		return cc, nil
	}

	cc, err := grpc.Dial(target, p.opts...)
	if err != nil {
		return nil, err
	}
	p.pool[target] = cc
	return cc, nil
}
