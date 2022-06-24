package debounce

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fired := make(chan struct{})
	d := New(func() error {
		close(fired)
		return errors.New("err")
	}, 50*time.Millisecond)

	c1 := d.Trigger(ctx)
	select {
	case <-fired:
		t.Fatal("didn't debounce")
	case <-time.After(30 * time.Millisecond):
	}
	c2 := d.Trigger(ctx)
	select {
	case <-fired:
		t.Fatal("didn't debounce")
	case <-time.After(30 * time.Millisecond):
	}
	c3 := d.Trigger(ctx)
	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("fn did not fire")
	}

	if err := <-c1; err == nil {
		t.Fatal("expected error")
	}
	if err := <-c2; err == nil {
		t.Fatal("expected error")
	}
	if err := <-c3; err == nil {
		t.Fatal("expected error")
	}
}

func TestMultipleCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var n int64
	d := New(func() error {
		atomic.AddInt64(&n, 1)
		return nil
	}, time.Millisecond)

	<-d.Trigger(ctx)
	<-d.Trigger(ctx)
	<-d.Trigger(ctx)
	if n != 3 {
		t.Fatalf("expected 3, got %d", n)
	}
}

func TestCancelCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := New(func() error {
		return nil
	}, time.Millisecond)

	if err := <-d.Trigger(ctx); err != context.Canceled {
		t.Fatalf("expected error context.Canceled, got %v", err)
	}
}
