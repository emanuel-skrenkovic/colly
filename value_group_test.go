package colly

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestHappy(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		vg := NewValueGroup[int](context.Background())

		for i := 0; i < 10; i++ {
			i := i
			vg.Go(func() (int, error) {
				return i, nil
			})
		}

		is, err := vg.Wait()

		if err != nil {
			t.Fail()
		}

		if len(is) < 10 {
			t.Fail()
		}
	}
}

func TestUnhappy(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		vg := NewValueGroup[int](context.Background())

		for i := 0; i < 10; i++ {
			i := i
			vg.Go(func() (int, error) {
				if i == 8 || i == 2 {
					return 0, fmt.Errorf("nope")
				}
				return i, nil
			})
		}

		is, err := vg.Wait()

		if err == nil {
			t.Fail()
		}

		if len(is) > 0 {
			t.Fail()
		}
	}
}

func TestContextTimeout(t *testing.T) {
	for i := 0; i < 1_000; i++ {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			vg := NewValueGroup[int](ctx)

			for i := 0; i < 10; i++ {
				i := i
				vg.Go(func() (int, error) {
					if i == 8 {
						time.Sleep(2 * time.Millisecond)
					}
					return i, nil
				})
			}

			is, err := vg.Wait()

			if err == nil {
				t.Error(err)
				t.Fail()
			}

			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("unexpected error type: %s", err.Error())
				t.Fail()
			}

			if len(is) > 0 {
				t.Errorf("too many responses")
				t.Fail()
			}
		}()
	}
}
