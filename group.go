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

	doneCh chan struct{}
	errCh  chan error

	errors []error
}

func NewGroup(ctx context.Context) *Group {
	return &Group{
		ctx:   ctx,
		wg:    sync.WaitGroup{},
		errCh: make(chan error),
	}
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		g.errCh <- f()
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
			if err != nil {
				g.errors = append(g.errors, err)
			}
			g.wg.Done()
		}
	}
}
