package colly

import (
	"context"
	"runtime"
	"sync"
)

type Group struct {
	ctx      context.Context
	wg       sync.WaitGroup
	capacity int
	actions  []func() error
	errCh    chan error
	values   []any
	errors   []error
}

type GroupOption func(*Group)

func WithMaxParallel(max int) GroupOption {
	return func(g *Group) {
		g.capacity = max
	}
}

func defaultCapacity() int { return runtime.GOMAXPROCS(0) }

func NewGroup(ctx context.Context, options ...GroupOption) *Group {
	g := Group{
		ctx:      ctx,
		wg:       sync.WaitGroup{},
		capacity: defaultCapacity(),
		errCh:    make(chan error, 1),
	}

	for _, opt := range options {
		opt(&g)
	}

	g.actions = make([]func() error, 0, g.capacity)
	g.reactorLoop()

	return &g
}

func (g *Group) Go(f func() error) error {
	g.actions = append(g.actions, f)
	go func() {
		g.wg.Add(1)
		g.errCh <- f()
	}()

	return nil
}

func (g *Group) Wait() {
	g.wg.Wait()
}

func (g *Group) reactorLoop() {
	for {
		select {
		case err := <-g.errCh:
			g.errors = append(g.errors, err)
		}
	}
}
