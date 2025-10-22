package goo

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"
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

func TestNewContextWithCancel(t *testing.T) {
	ctx, cancel := NewContext(context.Background()).WithCancel()

	fmt.Println("--1--", ctx.TraceId())

	go func() {
		fmt.Println("--2-- 暂停3秒", ctx.TraceId())
		time.Sleep(3 * time.Second)
		cancel()
	}()

	<-ctx.Done()

	fmt.Println("--3--", ctx.TraceId())
}

func TestNewContextWithTimeout(t *testing.T) {
	ctx, _ := NewContext(context.Background()).WithTimeout(3 * time.Second)

	fmt.Println("--1--", ctx.TraceId())

	select {
	case <-ctx.Done():
		fmt.Println("--2--", ctx.TraceId())
	}
	
	fmt.Println("--3--", ctx.TraceId())
}

func TestNewContextWithSignalNotify(t *testing.T) {
	ctx, _ := NewContext(context.Background()).WithSignalNotify()

	fmt.Println("--1--", ctx.TraceId())

	select {
	case <-ctx.Done():
		fmt.Println("--2--", ctx.TraceId())
	}

	fmt.Println("--3--", ctx.TraceId())
}
