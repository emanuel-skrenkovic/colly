package colly

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

// https://github.com/sourcegraph/conc

type Group struct {
	ctx context.Context
	wg  sync.WaitGroup

	capacity int

	doneCh chan struct{}
	errCh  chan error

	errors []error
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
		errCh:    make(chan error),
	}

	for _, opt := range options {
		opt(&g)
	}

	return &g
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		if err := f(); err != nil {
			g.errCh <- err
		}
	}()
}

func (g *Group) Wait() error {
	go g.reactorLoop()

	go func() {
		g.wg.Wait()
		close(g.doneCh)
	}()

	select {
	case <-g.doneCh:
		if len(g.errors) < 1 {
			return nil
		}

		var err error
		for _, newErr := range g.errors {
			err = errors.Join(err, newErr)
		}

		return err

	case <-g.ctx.Done():
		err := g.ctx.Err()

		for _, newErr := range g.errors {
			err = errors.Join(err, newErr)
		}

		return err
	}
}

func (g *Group) reactorLoop() {
	for {
		select {
		case err := <-g.errCh:
			g.errors = append(g.errors, err)
			g.wg.Done()
		}
	}
}
