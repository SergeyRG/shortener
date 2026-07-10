package pool

import (
	"sync"
)

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool sync.Pool
}

func New[T Resettable](new func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return new()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.pool.Put(item)
}
