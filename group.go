package colly

import (
	"context"
	"errors"
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

func NewGroup(ctx context.Context) *Group {
	return &Group{
		ctx:      ctx,
		wg:       sync.WaitGroup{},
		capacity: defaultCapacity(),
		errCh:    make(chan error),
	}
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
	go g.wg.Wait()
	go g.reactorLoop()

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
