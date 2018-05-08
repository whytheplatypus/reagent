package errgroup

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	g := &Group{}
	for i := 0; i < 10; i++ {
		g.Go(func(i int) func() error {
			return func() error {
				return fmt.Errorf("Stuff and things %d", i)
			}
		}(i))
	}
	err := g.Wait()
	if err == nil {
		t.Error("Expected an error to come back got nil")
	}
	t.Log(err)
}

func TestGroupCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	g := WithContext(ctx)
	for i := 0; i < 10; i++ {
		g.Go(func(i int) func() error {
			return func() error {
				return fmt.Errorf("Stuff and things %d", i)
			}
		}(i))
		time.Sleep(10 * time.Millisecond)
		if i == 4 {
			cancel()
		}
	}
	err := g.Wait()
	if err == nil {
		t.Error("Expected an error to come back got nil")
	}
	t.Log(err)
}
