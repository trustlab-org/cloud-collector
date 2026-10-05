package pull

import (
	"context"
	"sync"

	"github.com/warpstreamlabs/bento/public/service"
)

// Window tracks acknowledgement of every batch emitted for one poll window.
type Window struct {
	wg  sync.WaitGroup
	mu  sync.Mutex
	err error
}

func (w *Window) Track() service.AckFunc {
	w.wg.Add(1)
	var once sync.Once
	return func(_ context.Context, err error) error {
		once.Do(func() {
			if err != nil {
				w.mu.Lock()
				if w.err == nil {
					w.err = err
				}
				w.mu.Unlock()
			}
			w.wg.Done()
		})
		return nil
	}
}

// Wait blocks until every tracked batch is acked and returns the first nack error.
func (w *Window) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		w.mu.Lock()
		defer w.mu.Unlock()
		return w.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Pending is a batch waiting to be handed to Bento along with its ack.
type Pending struct {
	Batch service.MessageBatch
	Ack   service.AckFunc
}
