package colly

import (
	"context"
	"errors"
	"sync"
)

type ValueGroup[T any] struct {
	ctx context.Context
	wg  sync.WaitGroup

	maxParallel int

	errCh   chan error
	valueCh chan T
	doneCh  chan struct{}

	values []T
	errors []error
}

func NewValueGroup[T any](ctx context.Context) *ValueGroup[T] {
	return &ValueGroup[T]{
		ctx:     ctx,
		wg:      sync.WaitGroup{},
		errCh:   make(chan error),
		valueCh: make(chan T),
		doneCh:  make(chan struct{}),
	}
}

func (g *ValueGroup[T]) Go(f func() (T, error)) {
	g.wg.Add(1)

	go func() {
		res, err := f()
		if err != nil {
			g.errCh <- err
		} else {
			g.valueCh <- res
		}
	}()
}

func (g *ValueGroup[T]) Wait() ([]T, error) {
	go g.reactorLoop()

	go func() {
		g.wg.Wait()
		close(g.doneCh)
	}()

	select {
	case <-g.doneCh:
		if len(g.errors) < 1 {
			return g.values, nil
		}

		var err error
		for _, newErr := range g.errors {
			err = errors.Join(err, newErr)
		}

		return nil, err

	case <-g.ctx.Done():
		err := g.ctx.Err()

		for _, newErr := range g.errors {
			err = errors.Join(err, newErr)
		}

		return nil, err
	}
}

func (g *ValueGroup[T]) reactorLoop() {
	for {
		select {
		case err := <-g.errCh:
			g.errors = append(g.errors, err)
			g.wg.Done()
		case val := <-g.valueCh:
			g.values = append(g.values, val)
			g.wg.Done()
		}
	}
}
