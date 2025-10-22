package goo

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext(context.Background())

	var g errgroup.Group

	for i := range 10 {
		g.Go(func() error {
			ctx2 := NewContext(ctx).WithAppName(fmt.Sprintf("app-%d", i))
			ctx2.WithValue("index", i)
			fmt.Println(
				ctx2.TraceId(),
				ctx2.Value("index"),
				ctx2.TraceId(),
				ctx2.AppName(),
			)
			return nil
		})
	}

	g.Wait()
}
