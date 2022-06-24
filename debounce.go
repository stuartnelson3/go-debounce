package debounce

import (
	"context"
	"time"
)

func New(fn func() error, t time.Duration) *Debouncer {
	return &Debouncer{
		triggerc: make(chan chan<- error),
		timeout:  t,
		fn:       fn,
		active:   make(chan struct{}, 1),
	}
}

// Debouncer wraps a function fn that is to be debounced for the length of
// timeout. Additional calls to fire fn will reset the timer. After timeout
// elapses, fn is fired.
type Debouncer struct {
	triggerc chan chan<- error
	timeout  time.Duration
	fn       func() error

	// Used to show execution of fn is currently being debounced. Must be a
	// buffered channel with length 1.
	active chan struct{}
}

// Trigger sends a request to fire the function fn. a buffered channel which
// will contain the return value of fn is returned to the caller, which they
// are responsible for draining.
func (d *Debouncer) Trigger(ctx context.Context) <-chan error {
	res := make(chan error, 1)
	select {
	case d.active <- struct{}{}:
		// Update the internal state to show we're currently debouncing
		// fn execution and start the debounce method.
		go func() {
			d.debounce(ctx)
			// release lock on debouncing
			<-d.active
		}()
	default:
	}
	d.triggerc <- res
	return res
}

// debounce debounces function fn, so that repeated calls to trigger() will
// only execute fn a single time after timeout has elapsed. Repeated calls to
// trigger() reset the timer. Additionally, each call to trigger() sends a chan
// error on triggerc; the return value of fn is sent to all callers once fn has
// executed.
func (d *Debouncer) debounce(ctx context.Context) (err error) {
	t := time.NewTimer(d.timeout)
	callers := []chan<- error{}
	defer func() {
		for _, res := range callers {
			select {
			case res <- err:
				close(res)
			default:
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			return d.fn()
		case res := <-d.triggerc:
			callers = append(callers, res)
			if !t.Stop() {
				<-t.C
			}
			t.Reset(d.timeout)
		}
	}
}
