package goocontext

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestNew(t *testing.T) {
	ctx := New(context.Background()).WithValue("name", "hnatao")

	fmt.Println("--1--", ctx.TraceId())

	var g errgroup.Group

	for i := range 10 {
		g.Go(func() error {
			ctx2 := New(ctx).WithAppName("app-%d", i).WithValue("index", i)
			fmt.Println(
				ctx2.ValueString("name"),
				ctx2.TraceId(),
				ctx2.ValueInt("index"),
				ctx2.TraceId(),
				ctx2.AppName(),
			)
			return nil
		})
	}

	g.Wait()

	fmt.Println("--2--", ctx.TraceId())
}

func TestNewWithCancel(t *testing.T) {
	ctx, cancel := New(context.Background()).WithCancel()

	fmt.Println("--1--", ctx.TraceId())

	go func() {
		fmt.Println("--2-- 暂停3秒", ctx.TraceId())
		time.Sleep(3 * time.Second)
		cancel()
	}()

	<-ctx.Done()

	fmt.Println("--3--", ctx.TraceId())
}

func TestNewWithTimeout(t *testing.T) {
	ctx, _ := New(context.Background()).WithTimeout(3 * time.Second)

	fmt.Println("--1--", ctx.TraceId())

	<-ctx.Done()

	fmt.Println("--2--", ctx.TraceId())
}

func TestNewWithSignal(t *testing.T) {
	ctx := New(context.Background()).WithSignalNotify()

	fmt.Println("--1--", ctx.TraceId(), os.Getpid())

	<-ctx.Done()

	fmt.Println("--2--", ctx.TraceId())
}
