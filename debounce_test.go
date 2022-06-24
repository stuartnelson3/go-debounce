package debounce

import (
	"context"
	"errors"
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
