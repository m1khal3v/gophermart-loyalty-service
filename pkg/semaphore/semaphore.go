package semaphore

import "context"

type Semaphore struct {
	channel chan struct{}
}

func New(max uint64) *Semaphore {
	if max == 0 {
		panic("max cannot be 0")
	}

	return &Semaphore{
		channel: make(chan struct{}, max),
	}
}

func (semaphore *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case semaphore.channel <- struct{}{}:
		return nil
	}
}

func (semaphore *Semaphore) Release() {
	<-semaphore.channel
}
