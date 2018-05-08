package errgroup

import (
	"context"
	"fmt"
	"sync"
)

type SyncError struct {
	sync.Map
}

func (e *SyncError) Error() string {
	s := ""
	e.Range(func(key, value interface{}) bool {
		s = fmt.Sprintf("%s\n%s", s, key.(error).Error())
		return true
	})
	return s
}

func WithContext(ctx context.Context) *Group {
	return &Group{ctx: ctx}
}

type Group struct {
	ctx context.Context

	wg sync.WaitGroup

	errOnce sync.Once

	err error
}

func (g *Group) Wait() error {
	g.wg.Wait()
	return g.err
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		if g.ctx != nil && g.ctx.Err() != nil {
			return
		}

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = &SyncError{}
			})
			g.err.(*SyncError).Store(err, f)
		}
	}()
}
