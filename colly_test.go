package colly

import (
	"context"
	"testing"
)

func Test_Run(t *testing.T) {
	ctx := context.Background()
	g := NewGroup(ctx)

	err := g.Go(func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
